package internal

import "errors"

type Verse struct {
	DisplayName string  `json:"displayName"`
	Content     string  `json:"content"`
	Book        string  `json:"book"`
	Chapter     string  `json:"chapter"`
	Verse       string  `json:"verse"`
	Version     Version `json:"version"` //TODO versions
}

type Version string

const (
	AMP            Version = "amp"
	ASV            Version = "asv"
	CEV            Version = "cev"
	Darby          Version = "darby"
	ESV            Version = "esv"
	KJV            Version = "kjv"
	MSG            Version = "msg"
	NASB           Version = "nasb"
	NIV            Version = "niv"
	NKJV           Version = "nkjv"
	NLT            Version = "nlt"
	NRSV           Version = "nrsv"
	YLT            Version = "ylt"
	UnknownVersion Version = "unknown"
)

var allVersions = []Version{
	AMP,
	ASV,
	CEV,
	Darby,
	ESV,
	KJV,
	MSG,
	NASB,
	NIV,
	NKJV,
	NLT,
	NRSV,
	YLT,
}

func GetVersion(version string) (Version, error) {
	for _, v := range allVersions {
		if string(v) == version {
			return v, nil
		}
	}
	return UnknownVersion, errors.New("version does not exist")
}

type VerseService interface {
	Find(verse string, version Version) ([]Verse, error)
}
