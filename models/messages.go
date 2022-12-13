package models

type SMS struct {
	Payload    string `form:"payload" json:"payload" xml:"payload"  binding:"required"`
	Recipients string `form:"recipients" json:"recipients" xml:"recipients" binding:"required"`
	Originator string `form:"originator" json:"originator" xml:"originator" binding:"required"`
}

type MbRc struct {
	Payload    string `form:"payload" json:"payload" xml:"payload"  binding:"required"`
	Originator string `form:"originator" json:"originator" xml:"originator" binding:"required"`
}
