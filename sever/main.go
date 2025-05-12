package main

import (
	"bytes"
	"context"
	"fmt"
	"grpc-lesson/pb"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

func (*server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	fmt.Println("ListFiles was invoked")
	dir := "/Users/yusuke/dev/private/grpc-lesson/storage"

	paths, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, path := range paths {
		if !path.IsDir() {
			fileNames = append(fileNames, path.Name())
		}
	}

	res := &pb.ListFilesResponse{
		FileNames: fileNames,
	}

	return res, nil
}

func (*server) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	fmt.Println("DownloadFile was invoked")
	dir := "/Users/yusuke/dev/private/grpc-lesson/storage"
	fileName := req.GetFileName()
	filePath := filepath.Join(dir, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		res := &pb.DownloadFileResponse{
			FileData: buffer[:n],
		}
		sendErr := stream.Send(res)
		if sendErr != nil {
			return sendErr
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (*server) UploadFile(stream pb.FileService_UploadFileServer) error {
	fmt.Println("UploadFile was invoked")

	var buffer bytes.Buffer

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := &pb.UploadFileResponse{
				Size: int32(buffer.Len()),
			}
			return stream.SendAndClose(res)
		}
		if err != nil {
			return err
		}
		buffer.Write(req.GetFileData())
	}
}

func (*server) UploadAndNotifyProgress(stream pb.FileService_UploadAndNotifyProgressServer) error {
	fmt.Println("UploadAndNotifyProgress was invoked")

	size := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		data := req.GetFileData()
		log.Printf("Recieved data: %v", data)
		size += len(data)

		res := &pb.UploadAndNotifyProgressResponse{
			Message: fmt.Sprintf("Recieved %v bytes", size),
		}
		sendErr := stream.Send(res)
		if sendErr != nil {
			return sendErr
		}
	}
}

func main() {
	fmt.Println("Server is running...")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileServiceServer(grpcServer, &server{})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
