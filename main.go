package main

import (
	"flag"
	"os"

	"github.com/andriiyaremenko/ebayscraper/data"
	"github.com/andriiyaremenko/ebayscraper/requests"
	"github.com/andriiyaremenko/tinylog"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

var (
	login    string
	password string
	category string
	file     string
	logLevel int
)

func init() {
	flag.StringVar(&login, "login", "", "login for eBay authorization")
	flag.StringVar(&password, "password", "", "password for eBay authorization")
	flag.StringVar(&category, "category", "", "category to search in eBay")
	flag.StringVar(&file, "file", "./result.json", "file path to write scraping results to")
	flag.IntVar(&logLevel, "logLevel", 1, "log level")
	flag.Parse()
}

func main() {
	loggerFactory := tinylog.NewTinyLoggerFactory(os.Stderr)
	loggerFactory.SetLogLevel(logLevel)
	logger := loggerFactory.GetLogger(tinylog.ZeroModule)

	if login == "" {
		logger.Err("login is mandatory and should be set.")
		os.Exit(2)
	}
	if password == "" {
		logger.Err("password is mandatory and should be set.")
		os.Exit(2)
	}
	if category == "" {
		logger.Err("category is mandatory and should be set.")
		os.Exit(2)
	}

	c := colly.NewCollector()
	extensions.RandomUserAgent(c)
	extensions.Referer(c)
	logger.Info("started scraping...")

	ac := requests.NewAuthClient(loggerFactory)
	c = ac.CallEndpoint(c, login, password)

	fs := data.NewFileStorage(file)
	pc := requests.NewProductsClient(loggerFactory, fs)
	pc.CallEndpoint(c, category)

	logger.Info("finished scraping...")
}
