package core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/CycloneDX/cyclonedx-go"
	cdx "github.com/CycloneDX/cyclonedx-go"
	log "github.com/sirupsen/logrus"
)

const (
	JSON_FORMAT = "cyclonedxjson"
	XML_FORMAT  = "cyclonedxxml"
)

type CycloneDxManager struct {
	Option string
	Ext    string
	Format cyclonedx.BOMFileFormat
}

func NewCycloneDxManager(option string) (*CycloneDxManager, error) {
	ext, format, err := GetSbomExt(option)
	if err != nil {
		return nil, err
	}
	return &CycloneDxManager{
		Option: option,
		Ext:    ext,
		Format: format,
	}, nil
}

func GetSbomExt(option string) (string, cyclonedx.BOMFileFormat, error) {
	var format cdx.BOMFileFormat
	var ext string
	switch option {
	case JSON_FORMAT:
		ext = ".json"
		format = cdx.BOMFileFormatJSON
	case XML_FORMAT:
		ext = ".xml"
		format = cdx.BOMFileFormatXML
	default:
		return "", cdx.BOMFileFormatJSON, errors.New(fmt.Sprintf("Cyclonedx - Unknown presenter, type: %s", option))
	}

	return ext, format, nil
}

func (pres *CycloneDxManager) GetName(bom *cdx.BOM) (string, error) {

	var name string
	switch bom.Metadata.Component.Group {
	case "image":
		name = bom.Metadata.Component.Version
	case "directory":
		name = bom.Metadata.Component.Name
	default:
		return "", errors.New(fmt.Sprintf("Unknown sbom group %s", bom.Metadata.Component.Group))
	}

	return name, nil
}

func (pres *CycloneDxManager) Decode(reader io.Reader, bom *cdx.BOM) error {
	decoder := cdx.NewBOMDecoder(reader, pres.Format)

	if err := decoder.Decode(bom); err != nil {
		return err
	}

	return nil
}

func (pres *CycloneDxManager) Encode(writer io.Writer, bom *cdx.BOM) error {
	var encoder cdx.BOMEncoder

	encoder = cdx.NewBOMEncoder(writer, pres.Format)
	encoder.SetPretty(true)

	if err := encoder.Encode(bom); err != nil {
		return err
	}

	return nil
}

func (pres *CycloneDxManager) ReadFromFile(path string, bom *cdx.BOM) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	r := bufio.NewReader(f)
	return pres.Decode(r, bom)
}

func (pres *CycloneDxManager) WriteToFile(path string, bom *cdx.BOM) error {

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	defer w.Flush()

	log.Infof("Cyclonedx - Sbom pushed to FS, format: %s, Path: %s", pres.Option, path)
	return pres.Encode(w, bom)
}

func (pres *CycloneDxManager) WriteOut(output_list []string, bom *cdx.BOM) error {
	for _, path := range output_list {
		if path != "" {
			err := pres.WriteToFile(path, bom)
			if err != nil {
				log.Warnf("presenter - cyclonedx - write out fail, Path: %s", path)
				return err
			}
		}
	}

	return nil
}
