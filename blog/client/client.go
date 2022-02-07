package main

import (
	"blog/blogpb"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
	}

	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	//Create Blog
	blog := &blogpb.Blog{
		AuthorId: "x The Author",
		Title:    "Title X",
		Content:  "Content X",
	}

	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: blog,
	})
	if err != nil {
		log.Fatalf("Failed to create: %s", err)
	}

	log.Printf("Blog created: %s", res)

	readBlogRes, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		BlogId: res.Blog.GetId(),
	})
	if err != nil {
		log.Fatalf("Failed to get blog: %s", err)
	}

	log.Printf("Response :%s", readBlogRes)
}
