package models

import "cloud.google.com/go/firestore"

type Teacher struct {
	FirstName  string                   `firestore:"firstName" json:"firstName"`
	LastName   string                   `firestore:"lastName" json:"lastName"`
	StudentIDs []*firestore.DocumentRef `firestore:"studentIDs" json:"-"`
}
