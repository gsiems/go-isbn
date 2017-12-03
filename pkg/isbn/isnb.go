// Package isbn validates, parses, and converts International Standard Book Number (ISBN) data.
package isbn

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

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
	r, _ := regexp.Compile("[\\s-]")
	return r.ReplaceAllString(strings.ToUpper(isbn), "")
}

// chkLength checks that the length of the ISBN is correct
func chkLength(isbn string) (bool, error) {
	if len(isbn) != 10 && len(isbn) != 13 {
		return false, errors.New("ISBN length is incorrect.")
	}
	return true, nil
}

// chkCharacters checks that the characters in the ISBN are all valid
func chkCharacters(isbn string) (bool, error) {
	mainDigits := isbn[:len(isbn)-1]
	checkDigit := isbn[len(isbn)-1:]

	b := []byte(mainDigits)
	matched, err := regexp.Match("[^0-9]", b)
	if err != nil {
		return false, err
	} else if matched {
		return false, errors.New("Invalid character found in ISBN.")
	}

	b = []byte(checkDigit)
	matched, err = regexp.Match("[^0-9X]", b)
	if err != nil {
		return false, err
	} else if matched {
		return false, errors.New("Invalid character found in check digit.")
	}

	return true, nil
}

// chkCheckDigit checks that the check digit for the ISBN is correct
func chkCheckDigit(isbn string) (bool, error) {
	checkDigit := isbn[len(isbn)-1:]

	testDigit, err := calcCheckDigit(isbn)
	if err != nil {
		return false, err
	} else if testDigit != checkDigit {
		return false, errors.New("ISBN check digit is incorrect.")
	}

	return true, nil
}

// calcCheckDigit calculates the check digit for the ISBN
func calcCheckDigit(isbn string) (string, error) {
	if len(isbn) == 10 {
		return calcCheckDigit10(isbn)
	} else if len(isbn) == 13 {
		return calcCheckDigit13(isbn)
	}
	return "", errors.New("Unable to calculate check digit, ISBN is wrong length.")
}

// calcCheckDigit10 calculates the check digit for an ISBN-10
func calcCheckDigit10(isbn string) (string, error) {

	var cd string

	var sum int32 = 0
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

// calcCheckDigit10 calculates the check digit for an ISBN-13
func calcCheckDigit13(isbn string) (string, error) {

	var cd string
	var sum int32 = 0
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

// ParseISBN parses the supplied ISBN into its constituent elements and
// checks the validity of the ISBN.
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
		err = errors.New("No range data for parsing ISBNs (perhaps you did not LoadRangeData).")
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
		ret.Prefix = "978"
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
		chk, _ := calcCheckDigit10(isbn)
		ret.IsValid = ret.CheckDigit10 == chk

		// calculate the check digit for ISBN-13
		if ret.IsValid {
			ret.CheckDigit13, _ = calcCheckDigit13("978" + isbn)
		}
	} else if len(isbn) == 13 {
		ret.CheckDigit13 = isbn[len(isbn)-1:]
		chk, _ := calcCheckDigit13(isbn)
		ret.IsValid = ret.CheckDigit13 == chk

		// calculate the check digit for ISBN-10 if applicable
		if ret.IsValid && ret.Prefix == "978" {
			s := []string{ret.RegistrationGroup, ret.Registrant, ret.Publication, "X"}
			ret.CheckDigit10, _ = calcCheckDigit10(strings.Join(s, ""))
		}
	}

	return ret, nil
}
