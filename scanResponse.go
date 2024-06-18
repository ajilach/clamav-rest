package main

type scanResponse struct {
	Status      string `json:"Status"`
	Description string `json:"Description"`
	httpStatus  int    `json:"-"`
}
