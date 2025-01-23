//go:build !libpostal

package address

import (
	"strings"

	"github.com/moov-io/watchman/pkg/search"
	"github.com/moov-io/watchman/pkg/usaddress"
)

func ParseAddress(input string) search.Address {
	addr := usaddress.StandardizeAddress(input)

	// Construct line 1 from primary components
	line1Parts := []string{}

	// Handle PO Box
	if addr.POBox != "" {
		line1Parts = append(line1Parts, "PO Box "+addr.POBox)
	} else if addr.RuralRoute != "" {
		line1Parts = append(line1Parts, "RR "+addr.RuralRoute)
	} else if addr.HighwayContract != "" {
		line1Parts = append(line1Parts, "HC "+addr.HighwayContract)
	} else {
		// Standard street address
		if addr.PrimaryNumber != "" {
			line1Parts = append(line1Parts, addr.PrimaryNumber)
		}
		if addr.StreetPredir != "" {
			line1Parts = append(line1Parts, addr.StreetPredir)
		}
		if addr.StreetName != "" {
			line1Parts = append(line1Parts, addr.StreetName)
		}
		if addr.StreetSuffix != "" {
			line1Parts = append(line1Parts, addr.StreetSuffix)
		}
		if addr.StreetPostdir != "" {
			line1Parts = append(line1Parts, addr.StreetPostdir)
		}
	}

	return search.Address{
		Line1:      strings.Join(line1Parts, " "),
		Line2:      addr.SecondaryUnit,
		City:       addr.City,
		State:      addr.State,
		PostalCode: addr.ZIPCode + addr.Plus4,
		Country:    addr.Country,
	}
}
