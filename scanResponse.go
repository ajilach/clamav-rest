package main

type scanResponse struct {
	Status      string `json:"Status"`
	Description string `json:"Description"`
	FileName    string `json:"FileName,omitempty"`
	httpStatus  int    `json:"-"`
}
