package storage

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

const databasePath = "storage.db"

type content struct {
	Id   string `gorm:"column:id;primaryKey"`
	Text string `gorm:"column:text;not null" json:"text"`
}

func (content) TableName() string {
	return "content"
}

func (c *content) isNil() bool {
	return c.Id == "" && c.Text == ""
}

func NewStorage() (*Storage, error) {
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&content{}); err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) addContent(c *content) {
	s.db.FirstOrCreate(c)
}

const (
	url     = "https://uselessfacts.jsph.pl/random.json"
	timeout = time.Second
)

func (s *Storage) getFromRequest() (*content, error) {
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c := new(content)
	if err = json.Unmarshal(data, c); err != nil {
		return nil, err
	}

	if c.isNil() {
		return nil, errors.New("content is empty")
	}

	return c, nil
}

func (s *Storage) getFromDatabase() (*content, error) {
	c := new(content)
	if err := s.db.Order("RANDOM()").First(c).Error; err != nil {
		return nil, err
	}

	return c, nil
}

const defaultText = "This is default text!"

func (s *Storage) GetText() string {
	c, err := s.getFromRequest()
	if err == nil {
		s.addContent(c)
		return c.Text
	}

	c, err = s.getFromDatabase()
	if err == nil {
		return c.Text
	}

	return defaultText
}
