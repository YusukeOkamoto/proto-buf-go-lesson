package main

import (
	"context"
	"fmt"
	"grpc-lesson/pb"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Client is running...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	// callListFiles(client)
	// callDownloadFile(client)
	// callUploadFile(client)
	callUploadAndNotifyProgress(client)
}

func callListFiles(c pb.FileServiceClient) {
	res, err := c.ListFiles(context.Background(), &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalf("Failed to call ListFiles: %v", err)
	}

	fmt.Printf("Response: %v\n", res.GetFileNames())
}

func callDownloadFile(c pb.FileServiceClient) {
	req := &pb.DownloadFileRequest{
		FileName: "name.txt",
	}

	stream, err := c.DownloadFile(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to call DownloadFile: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive data: %v", err)
		}
		fmt.Printf("Response: %v\n", res.GetFileData())
		fmt.Println("--------------------------------")
		fmt.Printf("Response String: %v\n", string(res.GetFileData()))
	}
}

func callUploadFile(c pb.FileServiceClient) {
	filePath := "/Users/yusuke.okamoto/dev/study/proto-buf-go-lesson/storage/sports.txt"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stream, err := c.UploadFile(context.Background())
	if err != nil {
		log.Fatalf("Failed to call UploadFile: %v", err)
	}

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		req := &pb.UploadFileRequest{
			FileData: buffer[:n],
		}
		sendErr := stream.Send(req)
		if sendErr != nil {
			log.Fatalf("Failed to send file data: %v", sendErr)
		}
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to receive response: %v", err)
	}

	fmt.Printf("Received file size: %v\n", res.GetSize())
}

func callUploadAndNotifyProgress(c pb.FileServiceClient) {
	filePath := "/Users/yusuke.okamoto/dev/study/proto-buf-go-lesson/storage/sports.txt"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stream, err := c.UploadAndNotifyProgress(context.Background())
	if err != nil {
		log.Fatalf("Failed to call UploadAndNotifyProgress: %v", err)
	}

	// request側
	buf := make([]byte, 5)
	go func() {
		for {
			n, err := file.Read(buf)
			if n == 0 || err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Failed to read file: %v", err)
			}
			req := &pb.UploadAndNotifyProgressRequest{
				FileData: buf[:n],
			}
			sendErr := stream.Send(req)
			if sendErr != nil {
				log.Fatalf("Failed to send file data: %v", sendErr)
			}
			time.Sleep(2 * time.Second)
		}
		err := stream.CloseSend()
		if err != nil {
			log.Fatalf("Failed to close send: %v", err)
		}
	}()

	// response側
	ch := make(chan struct{})
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Failed to receive response: %v", err)
			}
			fmt.Printf("Response: %v\n", res.GetMessage())
		}
		close(ch)
	}()
	<-ch
}
