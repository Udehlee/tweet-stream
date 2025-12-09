package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/Udehlee/tweet-stream/gapi"
	"github.com/Udehlee/tweet-stream/internals/aggregator"
	"github.com/Udehlee/tweet-stream/internals/data/client"
	"github.com/Udehlee/tweet-stream/internals/data/generator"
	"github.com/Udehlee/tweet-stream/internals/data/simulated"
	"github.com/Udehlee/tweet-stream/internals/processor"
	"github.com/Udehlee/tweet-stream/internals/storage"
	"github.com/Udehlee/tweet-stream/models"
	"github.com/Udehlee/tweet-stream/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	influxDB, err := storage.ConnectToInfluxDB()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to InfluxDB")
	}
	defer influxDB.Close()

	generatedChan := make(chan *models.Tweet, 50)
	StreamChan := make(chan *models.Tweet, 50)

	tweetSvc := simulated.NewTweetService(&logger)
	cl := client.NewClient()
	gs := generator.NewGeneratorService(tweetSvc, cl, &logger, generatedChan)

	StartgRPCServer(generatedChan, &logger, ":50051")
	StartProcessor(ctx, "localhost:50051", StreamChan, &logger)

	go gs.GenerateTweets(ctx, 1*time.Second)

	agg := aggregator.NewTweetAggregator(generatedChan, influxDB, 5*time.Second)
	go agg.Start(ctx)

	logger.Info().Msg("Tweet streaming service running. Press Ctrl+C to exit.")
	select {}
}

func StartgRPCServer(tweetChan <-chan *models.Tweet, logger *zerolog.Logger, port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to listen on GRPC port")
	}

	streamServer := gapi.NewStreamServer(tweetChan)
	grpcServer := grpc.NewServer()
	pb.RegisterTweetServiceServer(grpcServer, streamServer)

	go func() {
		logger.Info().Msgf("GRPC server started on %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal().Err(err).Msg("Failed to serve GRPC")
		}
	}()
}

func StartProcessor(ctx context.Context, grpcTarget string, aggChan chan<- *models.Tweet, logger *zerolog.Logger) {
	processor := processor.NewProcessor(grpcTarget, aggChan)

	go func() {
		if err := processor.Run(ctx); err != nil {
			logger.Error().Err(err).Msg("StreamProcessor exited with error")
		}
	}()
}
