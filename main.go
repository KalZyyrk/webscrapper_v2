package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type (
	Player struct {
		Name string
		Pos  string
		Rank int
		Team string
	}
	playerJob struct {
		numPage int
		player  []Player
	}
)

const numPage = 7

func main() {
	start := time.Now()
	fName := "mlb_statcast.csv"
	file, err := os.Create(fName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()

	w.Write([]string{"Rank", "Player", "Pos", "Team"})

	var players []Player
	wg := sync.WaitGroup{}
	ch := make(chan playerJob, numPage)

	wg.Add(numPage)
	for i := 0; i < numPage; i++ {
		go GetData(ch, i+1)
	}

	go func() {
		for p := range ch {
			players = append(players, p.player...)
			wg.Done()
		}
	}()

	wg.Wait()

	close(ch)

	sort.SliceStable(players, func(i, j int) bool {
		return players[i].Rank < players[j].Rank
	})

	for _, p := range players {
		w.Write([]string{strconv.Itoa(p.Rank), p.Name, p.Pos, p.Team})
	}

	fmt.Println(time.Now().Sub(start))
}

func GetData(ch chan playerJob, numPage int) {
	var players []Player

	c := colly.NewCollector()

	c.OnHTML("tbody tr", func(e *colly.HTMLElement) {
		pName := fmt.Sprintf("%s %s", e.ChildText("div.top-wrapper-1NLTqKbE > div > a > span.full-3fV3c9pF:nth-child(1)"), e.ChildText("div.top-wrapper-1NLTqKbE > div > a > span.full-3fV3c9pF:nth-child(3)"))
		pRank, err := strconv.Atoi(e.ChildText("div.custom-cell-wrapper-34Cjf9P0 > div.index-3cdMSKi7"))
		if err != nil {
			panic(err)
		}
		player := Player{
			Rank: pRank,
			Pos:  e.ChildText("div.top-wrapper-1NLTqKbE > div.position-28TbwVOg"),
			Name: pName,
			Team: e.ChildText(`td[data-col="1"]`),
		}

		players = append(players, player)
	})

	url := fmt.Sprintf("https://www.mlb.com/stats?page=%d", numPage)

	c.Visit(url)

	ch <- playerJob{
		player: players,
	}
}
