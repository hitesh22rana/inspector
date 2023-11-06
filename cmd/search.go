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
	ErrorNoSearchPlatformChoosen = errors.New("please specify atleast one platform to search")
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
		webSites, err := getWebSitesList(username, false, false)
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

func getWebSitesList(username string, social bool, tech bool) ([]*WebSite, error) {
	var webSites []*WebSite

	if !social && !tech {
		return webSites, ErrorNoSearchPlatformChoosen
	}

	if social {
		webSites = append(webSites, getSocialWebsitesList(username)...)
	}

	if tech {
		webSites = append(webSites, getTechWebsitesList(username)...)
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
