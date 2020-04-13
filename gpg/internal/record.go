// Copyright (c) 2020 BVK Chaitanya

package internal

import (
	"log"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	version string
	fields  []string
}

func NewRecord(version string, fields []string) *Record {
	return &Record{version: version, fields: append([]string{}, fields...)}
}

func (r *Record) IsPublicKey() bool {
	return len(r.fields) > 0 && r.fields[0] == "pub"
}

func (r *Record) IsSecretKey() bool {
	return len(r.fields) > 0 && r.fields[0] == "sec"
}

func (r *Record) IsSecretSubkey() bool {
	return len(r.fields) > 0 && r.fields[0] == "ssb"
}

func (r *Record) IsUserID() bool {
	return len(r.fields) > 0 && r.fields[0] == "uid"
}

func (r *Record) IsFingerprint() bool {
	return len(r.fields) > 0 && r.fields[0] == "fpr"
}

func (r *Record) IsSubkey() bool {
	return len(r.fields) > 0 && r.fields[0] == "sub"
}

func (r *Record) KeyLength() int {
	if r.IsPublicKey() || r.IsSubkey() || r.IsSecretKey() || r.IsSecretSubkey() {
		if len(r.fields) >= 3 {
			size, err := strconv.Atoi(r.fields[2])
			if err != nil {
				log.Panicf("could not parse key length %q: %v", r.fields[2], err)
			}
			return size
		}
	}
	return -1
}

func (r *Record) KeyID() string {
	if r.IsPublicKey() || r.IsSubkey() || r.IsSecretKey() || r.IsSecretSubkey() {
		if len(r.fields) >= 5 {
			return r.fields[4]
		}
	}
	return ""
}

func (r *Record) KeyTrusted() bool {
	if r.IsPublicKey() || r.IsSubkey() || r.IsSecretKey() || r.IsSecretSubkey() {
		if len(r.fields) >= 2 {
			return strings.Contains(r.fields[1], "u")
		}
	}
	return false
}

func (r *Record) Fingerprint() string {
	if r.IsFingerprint() && len(r.fields) >= 10 {
		return r.fields[9]
	}
	return ""
}

func (r *Record) CanEncrypt() bool {
	if r.IsPublicKey() || r.IsSubkey() || r.IsSecretKey() || r.IsSecretSubkey() {
		return len(r.fields) >= 12 && strings.ContainsAny(r.fields[11], "e")
	}
	return false
}

func (r *Record) IsExpired() bool {
	if r.IsPublicKey() || r.IsSubkey() || r.IsSecretKey() || r.IsSecretSubkey() {
		return len(r.fields) >= 2 && strings.ContainsAny(r.fields[1], "e")
	}
	return false
}

func (r *Record) CreatedAt() time.Time {
	if r.IsPublicKey() || r.IsSubkey() || r.IsUserID() || r.IsSecretKey() || r.IsSecretSubkey() {
		if l := len(r.fields); l >= 6 && len(r.fields[5]) > 0 {
			field := r.fields[5]
			var secs, nsecs int64
			if p := strings.IndexRune(field, 'T'); p >= 0 {
				s, err := strconv.ParseInt(field[:p], 10, 64)
				if err != nil {
					log.Panicf("invalid seconds format in 5th field %q", field)
				}
				us, err := strconv.ParseInt(field[p+1:], 10, 32)
				if err != nil {
					log.Panicf("invalid microseconds format in 5th field %q", field)
				}
				secs, nsecs = s, us*1000
			} else {
				s, err := strconv.ParseInt(field, 10, 64)
				if err != nil {
					log.Panicf("invalid creation time format in 5th field %q", field)
				}
				secs = s
			}
			return time.Unix(secs, nsecs)
		}
	}
	return time.Time{}
}

func (r *Record) ExpiresAt() time.Time {
	if r.IsPublicKey() || r.IsSubkey() || r.IsUserID() || r.IsSecretKey() || r.IsSecretSubkey() {
		if l := len(r.fields); l >= 7 && len(r.fields[6]) > 0 {
			secs, err := strconv.ParseInt(r.fields[6], 10, 64)
			if err != nil {
				log.Panicf("invalid expiry time format in 7th field %q", r.fields[6])
			}
			return time.Unix(secs, 0)
		}
	}
	return time.Time{}
}

func (r *Record) UserDetail() string {
	if r.IsUserID() && len(r.fields) >= 10 {
		user := r.fields[9]
		return user
	}
	return ""
}

func (r *Record) UserName() string {
	detail := r.UserDetail()
	if p := strings.IndexAny(detail, "(<"); p >= 0 {
		return strings.TrimSpace(detail[:p])
	}
	return ""
}

func (r *Record) UserHash() string {
	if r.IsUserID() && len(r.fields) >= 8 {
		return r.fields[7]
	}
	return ""
}

func (r *Record) UserEmail() string {
	detail := r.UserDetail()
	if l := strings.LastIndexByte(detail, '<'); l >= 0 {
		if r := strings.LastIndexByte(detail, '>'); r >= l {
			return detail[l+1 : r]
		}
	}
	return ""
}
