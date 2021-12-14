package main

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/satmaelstorm/filup/cmd"
	"log"
	"strings"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Println(err)
	}
}

func main1() {
	endpoint := "localhost:9000"
	key := "minio"
	secret := "minio123"

	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(key, secret, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%#v\n", mc)
	ctx := context.Background()

	if b, err := mc.BucketExists(ctx, "test"); err != nil || !b {
		mc.MakeBucket(ctx, "test", minio.MakeBucketOptions{
			Region:        "eu-central-1",
			ObjectLocking: false,
		})
	}
	buf1 := bytes.NewBufferString(strings.Repeat("1", 1024*1024*6))
	mc.PutObject(ctx, "test", "t1", buf1, int64(buf1.Len()), minio.PutObjectOptions{
		ContentType: "text/plain",
	})

	buf2 := bytes.NewBufferString(strings.Repeat("2", 1024*1024*6))
	mc.PutObject(ctx, "test", "t2", buf2, int64(buf2.Len()), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	buf3 := bytes.NewBufferString(strings.Repeat("\nend\n", 1))
	mc.PutObject(ctx, "test", "t3", buf3, int64(buf3.Len()), minio.PutObjectOptions{
		ContentType: "text/plain",
	})

	ui, err := mc.ComposeObject(ctx, minio.CopyDestOptions{
		Bucket: "test",
		Object: "compile",
	},
		minio.CopySrcOptions{Bucket: "test", Object: "t1"},
		minio.CopySrcOptions{Bucket: "test", Object: "t2"},
		minio.CopySrcOptions{Bucket: "test", Object: "t3"},
	)

	log.Printf("%#v\n", ui)
	log.Println(err)
}
