// Пакет handlers, описание в файле doc.go
package handlers

import (
	"context"
	"errors"

	"log"

	"github.com/Niiazgulov/urlshortener-go.git/internal/configuration"
	pb "github.com/Niiazgulov/urlshortener-go.git/internal/handlers/proto"
	"github.com/Niiazgulov/urlshortener-go.git/internal/service"
	"github.com/Niiazgulov/urlshortener-go.git/internal/service/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const tokenHeader = "token"

// GRPC сервер
type Server struct {
	pb.UnimplementedURLShortenerServer
	repo    repository.AddorGetURL
	service service.ServiceStruct
	baseURL string
}

// Создание нового объекта структуры Server
func NewServer(repo repository.AddorGetURL, cfg *configuration.Config, service service.ServiceStruct) *Server {
	return &Server{
		repo:    repo,
		service: service,
		baseURL: cfg.BaseURLAddress,
	}
}

// Shorten получает запрос с URL-адресом и возвращает статус  и сокращенный URL
func (s *Server) Shorten(ctx context.Context, in *pb.ShortReq) (*pb.ShortResp, error) {
	userID, err := s.getCreateUserID(ctx)
	if err != nil {
		return nil, err
	}
	ourPoorURL := repository.URL{ShortURL: in.Url, UserID: userID}
	var resp pb.ShortResp
	shortid, handlerstatus, err := s.service.AddURL(ourPoorURL, userID)
	if err != nil {
		log.Println("Error while adding URl", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	resp.Status = pb.ShortStatus_CREATED
	if handlerstatus == 409 {
		resp.Status = pb.ShortStatus_UNKNOWN
	}
	resp.Result = s.baseURL + "/" + shortid

	return &resp, nil
}

// GetUserUrls возвращает список всех сокращенных URL пользователя
func (s *Server) GetUserUrls(ctx context.Context, _ *pb.UserUrlsReq) (*pb.UserUrlsResp, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unable to get userID")
	}

	var resp pb.UserUrlsResp
	urls, err := s.repo.FindAllUserUrls(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrKeyNotFound) {
		log.Println("Error while getting URLs", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	for urlID, originalURL := range urls {
		resp.Urls = append(resp.Urls,
			&pb.ShortOriginalUrl{
				ShortUrl:    s.baseURL + "/" + urlID,
				OriginalUrl: originalURL,
			})
	}

	return &resp, nil
}

// ShortBatch получает список URL/correlation_id и возвращает список сокращенных URL
func (s *Server) ShortBatch(ctx context.Context, in *pb.BatchReq) (*pb.BatchResp, error) {
	userID, err := s.getCreateUserID(ctx)
	if err != nil {
		return nil, err
	}

	var urls []repository.URL
	for _, v := range in.Urls {
		urls = append(urls, repository.URL{CorrelationID: v.CorrelationId, OriginalURL: v.OriginalUrl, UserID: userID})
	}

	corrIDUrlIDs, err := s.repo.BatchURL(ctx, urls)
	if err != nil {
		log.Println("Error while adding URls", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	var resp pb.BatchResp
	for _, v := range corrIDUrlIDs {
		resp.Urls = append(resp.Urls,
			&pb.CorrShort{
				CorrelationId: v.CorrelationID,
				ShortUrl:      s.baseURL + "/" + v.ShortURL,
			})
	}

	return &resp, nil
}

func (s *Server) getUserID(ctx context.Context) (string, error) {
	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(tokenHeader)
		if len(values) > 0 {
			token = values[0]
		}
	}

	return s.service.GetUserID(token)
}

func (s *Server) getCreateUserID(ctx context.Context) (string, error) {
	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(tokenHeader)
		if len(values) > 0 {
			token = values[0]
		}
	}

	userID, sign, err := s.service.GetCreateUserID(ctx, token)
	if err != nil {
		return "", status.Error(codes.Internal, "Internal server error")
	}

	header := metadata.New(map[string]string{tokenHeader: sign})
	if err := grpc.SendHeader(ctx, header); err != nil {
		return "", status.Errorf(codes.Internal, "Unable to send token header")
	}

	return userID, nil
}
