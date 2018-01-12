// Copyright 2017 Gregory Siems. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package isbn validates, parses, and converts International Standard Book Number (ISBN) data.
//
// In addition to simply checking the length, characters and check-digit
// of an ISBN this also parses the ISBN using the ISBN ranges data
// available from https://www.isbn-international.org/range_file_generation
// in order to verify that the Prefix, Registration Group, and Registrant
// elements of the ISBN are valid.
package isbn

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const p978 = "978"

// An ISBN contains the elements of a parsed ISBN. The elements
// consisting of:
//  [EAN.UCC prefix]-
//  [Registration Group element]-
//  [Registrant element]-
//  [Publication element]-
//  [Check-digit]
type ISBN struct {
	Prefix            string
	RegistrationGroup string
	Registrant        string
	Publication       string
	Agency            string
	CheckDigit10      string
	CheckDigit13      string
	IsValid           bool
}

// stripISBN removes all spaces and hypens from an ISBN. Since the
// ISBN to parse may, or may not, have spaces and/or hyphens we ensure
// that ther are none in order to simplify the parsing.
func stripISBN(isbn string) string {
	r, _ := regexp.Compile(`[\s-]`)
	return r.ReplaceAllString(strings.ToUpper(isbn), "")
}

// chkLength checks that the length of the ISBN is correct
func chkLength(isbn string) (bool, error) {
	if len(isbn) != 10 && len(isbn) != 13 {
		return false, errors.New("ISBN length is incorrect")
	}
	return true, nil
}

// chkCharacters checks that the characters in the ISBN are all valid.
func chkCharacters(isbn string) (bool, error) {
	mainDigits := isbn[:len(isbn)-1]
	checkDigit := isbn[len(isbn)-1:]

	b := []byte(mainDigits)
	matched, err := regexp.Match("[^0-9]", b)
	if err != nil {
		return false, err
	} else if matched {
		return false, errors.New("Invalid character found in ISBN")
	}

	b = []byte(checkDigit)
	matched, err = regexp.Match("[^0-9X]", b)
	if err != nil {
		return false, err
	} else if matched {
		return false, errors.New("invalid character found in check digit")
	}

	return true, nil
}

// chkCheckDigit checks that the check digit for the ISBN is correct.
func chkCheckDigit(isbn string) (bool, error) {
	checkDigit := isbn[len(isbn)-1:]

	_, err := chkLength(isbn)
	if err != nil {
		return false, err
	}

	var testDigit string

	if len(isbn) == 10 {
		testDigit, err = CalcCheckDigit10(isbn)
	} else if len(isbn) == 13 {
		testDigit, err = CalcCheckDigit13(isbn)
	}

	if err != nil {
		return false, err
	} else if testDigit != checkDigit {
		return false, errors.New("ISBN check digit is incorrect")
	}

	return true, nil
}

// CalcCheckDigit calculates the check digit for the ISBN.
func CalcCheckDigit(isbn string) (string, error) {

	isbn = stripISBN(isbn)

	_, err := chkLength(isbn)
	if err != nil {
		return "", err
	}

	_, err = chkCharacters(isbn)
	if err != nil {
		return "", err
	}

	if len(isbn) == 10 {
		return CalcCheckDigit10(isbn)
	}
	return CalcCheckDigit13(isbn)
}

// CalcCheckDigit10 calculates the check digit for an ISBN-10.
func CalcCheckDigit10(isbn string) (string, error) {

	var cd string

	var sum int32
	var i int32
	for i = 0; i < 9; i++ {
		v1, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return cd, err
		}
		sum += (10 - i) * int32(v1)
	}
	rem := (11 - sum%11) % 11
	if rem == 10 {
		cd = "X"
	} else {
		cd = strconv.Itoa(int(rem))
	}

	return cd, nil
}

// CalcCheckDigit13 calculates the check digit for an ISBN-13.
func CalcCheckDigit13(isbn string) (string, error) {

	var cd string
	var sum int32
	var i int32
	for i = 0; i < 12; i += 2 {
		v1, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return cd, err
		}
		v2, err := strconv.Atoi(string(isbn[i+1]))
		if err != nil {
			return cd, err
		}
		sum += int32(v1) + 3*int32(v2)

	}
	rem := (10 - sum%10) % 10
	cd = strconv.Itoa(int(rem))

	return cd, nil
}

// ValidateCheckDigit test whether or not the check digit for an ISBN
// matches the calculated check digit.
func ValidateCheckDigit(isbn string) bool {
	isbn = stripISBN(isbn)
	provided := isbn[len(isbn)-1:]

	calculated, _ := CalcCheckDigit(isbn)

	return provided == calculated
}

// ParseISBN parses the supplied ISBN into its constituent elements and
// checks the validity of the elements.
func ParseISBN(isbn string) (ISBN, error) {

	var ret ISBN

	isbn = stripISBN(isbn)

	// Ensure that the basic format (length, characters) of the ISBN
	// match the specification.
	_, err := chkLength(isbn)
	if err != nil {
		return ret, err
	}

	_, err = chkCharacters(isbn)
	if err != nil {
		return ret, err
	}

	_, err = chkCheckDigit(isbn)
	if err != nil {
		return ret, err
	}

	// Ensure that the range data has been loaded so that the ISBN can
	// be parsed and that the remainder of the validation can be
	// performed.
	if !HasRangeData() {
		err = errors.New("no range data for parsing ISBNs (perhaps you did not LoadRangeData)")
		return ret, err
	}

	// Since the different elements of an ISBN are of variable length it
	// is necessary to start peeling one character at a time from the
	// left hand side of the ISBN until the next needed element can be
	// matched (or the end of the ISBN is reached without a match).
	// https://en.wikipedia.org/wiki/Isbn has a reasonable writeup on
	// the subject.
	//
	// https://www.isbn-international.org/range_file_generation#
	//  RangeMessage.xml
	//
	// For example:
	//   8804473282 parses as:
	//       -        | Prefix
	//       88-      | RegistrationGroup (Italy)
	//       04-      | Registrant (Mondadori)
	//       47328-   | Publication (Attenti al gorilla)
	//       2        | Check digit
	//
	//   9780670013951 parses as: ISBN 88-04-47328-2
	//       978-     | Prefix
	//       0-       | RegistrationGroup (English language)
	//       670-     | Registrant (Viking Books for Young Readers)
	//       01395-   | Publication (Llama Llama and the Bully Goat)
	//       1        | Check digit

	if len(isbn) == 10 {
		ret.Prefix = p978
	}

	rg := []byte(isbn[:len(isbn)-1])
	var pfx []byte
	var grp []byte
	var reg []byte
	var pub []byte

	var rs registrant

	for _, digit := range rg {

		if ret.Prefix == "" {
			pfx = append(pfx, digit)
			_, ok := rmd[string(pfx[:])]
			if ok {
				ret.Prefix = string(pfx[:])
			}
		} else if ret.RegistrationGroup == "" {
			grp = append(grp, digit)
			_, ok := rmd[ret.Prefix][string(grp[:])]
			if ok {
				ret.RegistrationGroup = string(grp[:])
				rs = rmd[ret.Prefix][string(grp[:])]
				ret.Agency = rs.Agency
			}
		} else if ret.Registrant == "" {
			reg = append(reg, digit)
			chk, err := strconv.Atoi(string(reg[:]))
			if err != nil {
				return ret, err
			}

			for _, rg := range rs.Ranges {
				rStart := rg[0]
				rEnd := rg[1]
				rLen := rg[2]
				if len(reg) == rLen && chk >= rStart && chk <= rEnd {
					ret.Registrant = string(reg[:])
				}
			}
		} else {
			pub = append(pub, digit)
			ret.Publication = string(pub[:])
		}
	}

	// Check the check digit
	if len(isbn) == 10 {
		ret.CheckDigit10 = isbn[len(isbn)-1:]
		chk, _ := CalcCheckDigit10(isbn)
		ret.IsValid = ret.CheckDigit10 == chk

		// calculate the check digit for ISBN-13
		if ret.IsValid {
			ret.CheckDigit13, _ = CalcCheckDigit13(p978 + isbn)
		}
	} else if len(isbn) == 13 {
		ret.CheckDigit13 = isbn[len(isbn)-1:]
		chk, _ := CalcCheckDigit13(isbn)
		ret.IsValid = ret.CheckDigit13 == chk

		// calculate the check digit for ISBN-10 if applicable
		if ret.IsValid && ret.Prefix == p978 {
			s := []string{ret.RegistrationGroup, ret.Registrant, ret.Publication, "X"}
			ret.CheckDigit10, _ = CalcCheckDigit10(strings.Join(s, ""))
		}
	}

	return ret, nil
}

// ISBN13 returns the ISBN as an ISBN-13.
func (x ISBN) ISBN13() string {
	if x.IsValid {
		s13 := []string{x.Prefix, x.RegistrationGroup, x.Registrant, x.Publication, x.CheckDigit13}
		return strings.Join(s13, "")
	}
	return ""
}

// ISBN10 returns the ISBN as an ISBN-10 (as applicable).
func (x ISBN) ISBN10() string {
	if x.Prefix == p978 {
		s10 := []string{x.RegistrationGroup, x.Registrant, x.Publication, x.CheckDigit10}
		return strings.Join(s10, "")
	}
	return ""
}

// String implements the Stringer interface. Format currently subject to change.
func (x ISBN) String() string {
	if x.IsValid {
		s13 := []string{x.Prefix, x.RegistrationGroup, x.Registrant, x.Publication, x.CheckDigit13}
		out := strings.Join(s13, "-")

		if x.Prefix == p978 {
			s10 := []string{x.RegistrationGroup, x.Registrant, x.Publication, x.CheckDigit10}
			out = out + " (" + strings.Join(s10, "-") + ")"
		}
		return out
	}
	return ""
}
