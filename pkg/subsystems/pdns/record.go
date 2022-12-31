package pdns

import (
	"context"
	"errors"
	"fmt"
	"github.com/joeig/go-powerdns/v3"
	"go-deploy/pkg/subsystems/pdns/models"
	"strings"
)

func getFullHostname(zone, hostName string) string {
	// trailing dot is intentional
	name := fmt.Sprintf("%s.%s.", hostName, zone)
	return name
}

func createID(zone, hostname, recordType string) string {
	id := fmt.Sprintf("%s_%s_%s", zone, hostname, recordType)
	return id
}

type parsedRecordID struct {
	Zone       string
	Hostname   string
	RecordType string
}

func parseID(id string) *parsedRecordID {
	splits := strings.Split(id, "_")
	if len(splits) != 3 {
		return nil
	}

	return &parsedRecordID{
		Zone:       splits[0],
		Hostname:   splits[1],
		RecordType: splits[2],
	}
}

func (client *Client) ReadRecord(id string) (*models.RecordPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read record %s. details: %s", id, err)
	}

	idParsed := parseID(id)

	zone, err := client.PdnsClient.Zones.Get(context.TODO(), idParsed.Zone)
	if err != nil {
		return nil, makeError(err)
	}

	if zone == nil {
		return nil, nil
	}

	fullHostname := getFullHostname(idParsed.Zone, idParsed.Hostname)

	for _, rrset := range zone.RRsets {
		if rrset.Name == nil || rrset.Type == nil || rrset.TTL == nil {
			continue
		}

		if *rrset.Name == fullHostname && len(rrset.Records) > 0 {
			var content []string
			for _, record := range rrset.Records {
				if record.Content != nil {
					content = append(content, *record.Content)
				}
			}

			hostname := strings.TrimSuffix(*rrset.Name, idParsed.Zone)

			return &models.RecordPublic{
				ID:         id,
				Hostname:   hostname,
				RecordType: string(*rrset.Type),
				TTL:        *rrset.TTL,
				Content:    content,
			}, nil
		}
	}

	return nil, nil
}

func (client *Client) CreateRecord(public *models.RecordPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create record %s. details: %s", public.Hostname, err)
	}

	if public.Hostname == "" {
		return "", makeError(errors.New("hostname required"))
	}

	fullHostname := getFullHostname(client.Zone, public.Hostname)

	err := client.PdnsClient.Records.Add(context.TODO(), client.Zone, fullHostname, powerdns.RRType(public.RecordType), 60, public.Content)
	if err != nil {
		return "", err
	}

	id := createID(client.Zone, public.Hostname, public.RecordType)

	return id, nil
}

func (client *Client) DeleteRecord(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete record %s. details: %s", id, err)
	}

	idParsed := parseID(id)

	fullHostname := getFullHostname(idParsed.Zone, idParsed.Hostname)

	err := client.PdnsClient.Records.Delete(context.TODO(), client.Zone, fullHostname, powerdns.RRType(idParsed.RecordType))
	if err != nil {
		return makeError(err)
	}

	return nil
}
