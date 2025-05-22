package models

import "cloud.google.com/go/firestore"

type Student struct {
	FirstName  string                   `firestore:"firstName" json:"firstName"`
	LastName   string                   `firestore:"lastName" json:"lastName"`
	TeacherIDs []*firestore.DocumentRef `firestore:"teacherIDs" json:"-"`
}
