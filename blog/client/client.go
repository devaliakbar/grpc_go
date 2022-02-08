package main

import (
	"blog/blogpb"
	"context"
	"io"
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

	//Read Blog
	readBlogRes, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		BlogId: res.Blog.GetId(),
	})
	if err != nil {
		log.Fatalf("Failed to get blog: %s", err)
	}

	log.Printf("Response :%s", readBlogRes)

	//UPDATE BLOG
	updateBlog := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       res.Blog.GetId(),
			AuthorId: "Change Author",
			Title:    "Title Changed",
			Content:  "Content Changed",
		},
	}

	updateRes, upRrr := c.UpdateBlog(context.Background(), updateBlog)
	if upRrr != nil {
		log.Fatalf("Error happened while updating: %s", upRrr)
	}

	log.Printf("Update Response: %s", updateRes)

	//DELETE BLOG
	dltBlog, dltErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{
		BlogId: res.Blog.GetId(),
	})

	if dltErr != nil {
		log.Fatalf("Failed to delete: %s", dltErr)
	}

	log.Printf("Blog deleted: %s", dltBlog)

	///List Blogs
	stream, listBlogerr := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})

	if listBlogerr != nil {
		log.Fatalf("Failed to list blogs: %s", listBlogerr)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive: %s", res)
		}

		log.Printf("Received from list blog request: %s", res)
	}
}
