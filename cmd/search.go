/*
Copyright © 2023 Hitesh Rana hitesh22rana@gmail.com
*/
package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
	wg                           sync.WaitGroup
	maxProcs                     = runtime.NumCPU()
	ErrorNoUsernameProvided      = errors.New("please specify a username to search")
	ErrorNoSearchPlatformChoosen = errors.New("please specify atleast one platform to search")
)

// function to validate platform flags
func validatePlatformFlag(platformFlag []string) error {
	if len(platformFlag) < 1 {
		return ErrorNoSearchPlatformChoosen
	}

	for _, value := range platformFlag {
		switch value {
		case "social":
			continue
		case "tech":
			continue
		default:
			return fmt.Errorf("invalid platform %s", value)
		}
	}
	return nil
}

// searchCmd represents the username command
var searchCmd = &cobra.Command{
	Use:   "search [username]",
	Short: "Searches for the given username on different platforms.",
	Long:  `Search (inspector search) will search for the provided username on different platforms.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cobra.CheckErr(ErrorNoUsernameProvided)
		}

		platforms, err := cmd.Flags().GetStringSlice("platform")
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		err = validatePlatformFlag(platforms)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		username := args[0]
		webSites, err := getWebSitesList(username, platforms)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		searchMatches := Search(webSites)

		if len(searchMatches) < 1 {
			fmt.Println("No matches found.")
			return
		}

		fmt.Printf("\nusername: %s was found on:-\n", username)
		for _, match := range searchMatches {
			fmt.Printf("%s : %s\n", match.name, match.url)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringSliceP("platform", "p", []string{"social"}, "Platform(s) to search users on")
}

type WebSite struct {
	name  string
	url   string
	check func(res *http.Response) bool
}

type SearchResult struct {
	name string
	url  string
}

func getSocialWebsitesList(username string) []*WebSite {
	webSites := []*WebSite{
		{
			name: "LinkedIn",
			url:  fmt.Sprintf("https://www.linkedin.com/in/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
		{
			name: "Facebook",
			url:  fmt.Sprintf("https://www.facebook.com/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
		{
			name: "Instagram",
			url:  fmt.Sprintf("https://www.instagram.com/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
		{
			name: "Twitter",
			url:  fmt.Sprintf("https://twitter.com/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
		{
			name: "BioLink",
			url:  fmt.Sprintf("https://bio.link/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
	}

	return webSites
}

func getTechWebsitesList(username string) []*WebSite {
	webSites := []*WebSite{
		{
			name: "LeetCode",
			url:  fmt.Sprintf("https://leetcode.com/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
		{
			name: "Github",
			url:  fmt.Sprintf("https://github.com/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
		{
			name: "Codeforces",
			url:  fmt.Sprintf("https://codeforces.com/profile/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
		{
			name: "HackerEarth",
			url:  fmt.Sprintf("https://www.hackerearth.com/@%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
		{
			name: "Codechef",
			url:  fmt.Sprintf("https://www.codechef.com/users/%s", username),
			check: func(res *http.Response) bool {
				return res.StatusCode == 200
			},
		},
	}

	return webSites
}

func getWebSitesList(username string, platforms []string) ([]*WebSite, error) {
	var webSites []*WebSite

	for _, platform := range platforms {
		switch platform {
		case "social":
			webSites = append(webSites, getSocialWebsitesList(username)...)
		case "tech":
			webSites = append(webSites, getTechWebsitesList(username)...)
		}
	}

	return webSites, nil

}

func Search(webSites []*WebSite) []*SearchResult {
	var searchMatches []*SearchResult

	limiter := make(chan bool, maxProcs)
	resultChan := make(chan *SearchResult, len(webSites))

	for _, webSite := range webSites {
		limiter <- true
		wg.Add(1)

		go func(site *WebSite) {
			defer func() {
				<-limiter
				wg.Done()
			}()

			res, err := client.Get(site.url)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer res.Body.Close()

			if ok := site.check(res); !ok {
				fmt.Printf("[-] user was Not found on %s\n", site.name)
				return
			}

			fmt.Printf("[-_0] user was found on %s\n", site.name)
			resultChan <- &SearchResult{
				name: site.name,
				url:  site.url,
			}
		}(webSite)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		searchMatches = append(searchMatches, result)
	}

	return searchMatches
}
