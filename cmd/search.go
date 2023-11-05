/*
Copyright © 2023 Hitesh Rana hitesh22rana@gmail.com
*/
package cmd

import (
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
	wg       sync.WaitGroup
	maxProcs = runtime.NumCPU()
)

// searchCmd represents the username command
var searchCmd = &cobra.Command{
	Use:   "search [username]",
	Short: "Searches for the given username on different platforms.",
	Long:  `Search (inspector search) will search for the provided username on different platforms.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var comps []string

		if len(args) == 0 {
			comps = cobra.AppendActiveHelp(comps, "Please specify the username to search")
		} else if len(args) == 1 {
			comps = cobra.AppendActiveHelp(comps, "This command does not take any more arguments (but may accept flags)")
		} else {
			comps = cobra.AppendActiveHelp(comps, "ERROR: Too many arguments specified")
		}
		return comps, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cobra.CheckErr(fmt.Errorf("search needs a username to search"))
		}

		username := args[0]
		webSites := getWebSitesList(username)
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

func getWebSitesList(username string) []*WebSite {
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
	}

	return webSites
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
