package pdfcpu

import (
	"github.com/hhrutter/pdfcpu/pkg/log"
	"github.com/pkg/errors"
)

func memberOf(s string, list []string) bool {

	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

func intMemberOf(i int, list []int) bool {
	for _, v := range list {
		if i == v {
			return true
		}
	}
	return false
}

func validateCreationDate(xRefTable *XRefTable, o PDFObject) (err error) {

	if xRefTable.ValidationMode == ValidationRelaxed {
		_, err = validateString(xRefTable, o, nil)
	} else {
		_, err = validateDateObject(xRefTable, o, V10)
	}

	return err
}

func handleDefault(xRefTable *XRefTable, o PDFObject) (err error) {

	if xRefTable.ValidationMode == ValidationStrict {
		_, err = xRefTable.DereferenceStringOrHexLiteral(o, V10, nil)
	} else {
		_, err = xRefTable.Dereference(o)
	}

	return err
}

func validateDocumentInfoDict(xRefTable *XRefTable, obj PDFObject) (hasModDate bool, err error) {

	// Document info object is optional.

	dict, err := xRefTable.DereferenceDict(obj)
	if err != nil || dict == nil {
		return false, err
	}

	for k, v := range dict.Dict {

		switch k {

		// text string, opt, since V1.1
		case "Title":
			_, err = xRefTable.DereferenceStringOrHexLiteral(v, V11, nil)

		// text string, optional
		case "Author":
			_, err = xRefTable.DereferenceStringOrHexLiteral(v, V10, nil)

		// text string, optional, since V1.1
		case "Subject":
			_, err = xRefTable.DereferenceStringOrHexLiteral(v, V11, nil)

		// text string, optional, since V1.1
		case "Keywords":
			_, err = xRefTable.DereferenceStringOrHexLiteral(v, V11, nil)

		// text string, optional
		case "Creator":
			_, err = xRefTable.DereferenceStringOrHexLiteral(v, V10, nil)

		// text string, optional
		case "Producer":
			_, err = xRefTable.DereferenceStringOrHexLiteral(v, V10, nil)

		// date, optional
		case "CreationDate":
			err = validateCreationDate(xRefTable, v)

		// date, required if PieceInfo is present in document catalog.
		case "ModDate":
			hasModDate = true
			_, err = validateDateObject(xRefTable, v, V10)

		// name, optional, since V1.3
		case "Trapped":
			validate := func(s string) bool { return memberOf(s, []string{"True", "False", "Unknown"}) }
			_, err = xRefTable.DereferenceName(v, V13, validate)

		// text string, optional
		default:
			err = handleDefault(xRefTable, v)

		}

		if err != nil {
			return false, err
		}

	}

	return hasModDate, nil
}

func validateDocumentInfoObject(xRefTable *XRefTable) error {

	// Document info object is optional.
	if xRefTable.Info == nil {
		return nil
	}

	log.Debug.Println("*** validateDocumentInfoObject begin ***")

	hasModDate, err := validateDocumentInfoDict(xRefTable, *xRefTable.Info)
	if err != nil {
		return err
	}

	hasPieceInfo, err := xRefTable.CatalogHasPieceInfo()
	if err != nil {
		return err
	}

	if hasPieceInfo && !hasModDate {
		return errors.Errorf("validateDocumentInfoObject: missing required entry \"ModDate\"")
	}

	log.Debug.Println("*** validateDocumentInfoObject end ***")

	return nil
}
