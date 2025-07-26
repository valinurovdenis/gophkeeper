package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/valinurovdenis/gophkeeper/client/client"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func AddCommands(rootCmd *cobra.Command) {

}

func Execute() {
	client, err := client.NewGophKeeperClient()

	var (
		filePath string
		fileId   string
		login    string
		password string
		fileName string
		comment  string
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var rootCmd = &cobra.Command{
		Use:   "gophkeeper",
		Short: "Gophkeeper file manager",
	}

	var downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download file with given id to local path",
		Run: func(cmd *cobra.Command, args []string) {
			client.DownloadFile(context.Background(), filePath, fileId)
		},
	}
	downloadCmd.Flags().StringVar(&filePath, "path", "", "local path")
	downloadCmd.Flags().StringVar(&fileId, "id", "", "file id")

	var uploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload file with given path",
		Run: func(cmd *cobra.Command, args []string) {
			client.UploadFile(context.Background(), filePath, fileName, comment)
		},
	}
	uploadCmd.Flags().StringVar(&filePath, "path", "", "local path")
	uploadCmd.Flags().StringVar(&comment, "comment", "", "file comment")
	uploadCmd.Flags().StringVar(&fileName, "name", "", "file name")

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete file with given id",
		Run: func(cmd *cobra.Command, args []string) {
			client.DeleteFile(context.Background(), fileId)
		},
	}
	deleteCmd.Flags().StringVar(&fileId, "id", "", "file id")

	var registerCmd = &cobra.Command{
		Use:   "register",
		Short: "Register user",
		Run: func(cmd *cobra.Command, args []string) {
			client.Register(context.Background(), login, password)
		},
	}
	registerCmd.Flags().StringVar(&login, "login", "", "user login")
	registerCmd.Flags().StringVar(&password, "password", "", "user password")

	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login user",
		Run: func(cmd *cobra.Command, args []string) {
			client.Login(context.Background(), login, password)
		},
	}
	loginCmd.Flags().StringVar(&login, "login", "", "user login")
	loginCmd.Flags().StringVar(&password, "password", "", "user password")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Build version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Build version: %s\n", buildVersion)
			fmt.Printf("Build date: %s\n", buildDate)
			fmt.Printf("Build commit: %s\n", buildCommit)
		},
	}

	var listFilesCmd = &cobra.Command{
		Use:   "list-files",
		Short: "List user files",
		Run: func(cmd *cobra.Command, args []string) {
			client.ListFiles(context.Background())
		},
	}

	rootCmd.AddCommand(downloadCmd, uploadCmd, deleteCmd, registerCmd, loginCmd, listFilesCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
