package download

import (
	"context"
	"fmt"
	"time"

	"github.com/moov-io/watchman/pkg/ofac"
	"github.com/moov-io/watchman/pkg/search"

	"github.com/moov-io/base/log"
	"golang.org/x/sync/errgroup"
)

type Downloader interface {
	RefreshAll(ctx context.Context) (Stats, error)
}

func NewDownloader(logger log.Logger, conf Config) (Downloader, error) {
	return &downloader{
		logger: logger,
		conf:   conf,
	}, nil
}

type downloader struct {
	logger log.Logger
	conf   Config
}

func (dl *downloader) RefreshAll(ctx context.Context) (Stats, error) {
	stats := Stats{
		Lists:     make(map[string]int),
		StartedAt: time.Now().In(time.UTC),
	}

	logger := dl.logger.Info().With(log.Fields{
		"initial_data_directory": log.String(dl.conf.InitialDataDirectory),
	})
	start := time.Now()
	logger.Info().Log("starting list refresh")

	g, ctx := errgroup.WithContext(ctx)
	preparedLists := make(chan preparedList, 1)

	fmt.Println("starting all lists")

	g.Go(func() error {
		err := loadOFACRecords(ctx, logger, dl.conf, preparedLists)
		if err != nil {
			return fmt.Errorf("loading OFAC records: %w", err)
		}
		return nil
	})

	err := g.Wait()
	close(preparedLists)
	fmt.Printf("finished all lists: %v\n", time.Since(start))
	start = time.Now()

	if err != nil {
		return stats, fmt.Errorf("problem loading lists: %v", err)
	}

	// accumulate the lists
	fmt.Println("starting to accumulate preparedLists")
	for list := range preparedLists {
		stats.Lists[string(list.ListName)] = len(list.Entities)
		stats.Entities = append(stats.Entities, list.Entities...)
	}
	fmt.Printf("finished accumulating preparedLists: %v\n", time.Since(start))

	stats.EndedAt = time.Now().In(time.UTC)

	return stats, nil
}

type preparedList struct {
	ListName search.SourceList
	Entities []search.Entity[search.Value]
}

func loadOFACRecords(ctx context.Context, logger log.Logger, conf Config, responseCh chan preparedList) error {
	start := time.Now()
	fmt.Println("starting OFAC download")
	files, err := ofac.Download(ctx, logger, conf.InitialDataDirectory)
	fmt.Printf("finished OFAC download: %v\n", time.Since(start))
	start = time.Now()
	if err != nil {
		return fmt.Errorf("download: %v", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("unexpected %d OFAC files found", len(files))
	}

	fmt.Println("starting OFAC parse")
	res, err := ofac.Read(files)
	fmt.Printf("finished OFAC parse: %v\n", time.Since(start))
	start = time.Now()
	if err != nil {
		return err
	}

	fmt.Println("starting OFAC grouping")
	entities := ofac.GroupIntoEntities(res.SDNs, res.Addresses, res.SDNComments, res.AlternateIdentities)
	fmt.Printf("finished OFAC grouping: %v\n", time.Since(start))
	start = time.Now()

	fmt.Println("sending OFAC preparedList")
	responseCh <- preparedList{
		ListName: search.SourceUSOFAC,
		Entities: entities,
	}
	fmt.Printf("finished sending OFAC preparedList: %v\n", time.Since(start))

	return nil
}

// 	"github.com/moov-io/watchman/pkg/csl_eu"
// 	"github.com/moov-io/watchman/pkg/csl_uk"
// 	"github.com/moov-io/watchman/pkg/csl_us"

// func cslUSRecords(logger log.Logger, initialDir string) (csl_us.CSL, error) {
// 	file, err := csl_us.Download(logger, initialDir)
// 	if err != nil {
// 		logger.Warn().Logf("skipping CSL US download: %v", err)
// 		return csl_us.CSL{}, nil
// 	}
// 	cslRecords, err := csl_us.ReadFile(file["csl.csv"])
// 	if err != nil {
// 		return csl_us.CSL{}, fmt.Errorf("reading CSL US: %w", err)
// 	}
// 	return cslRecords, nil
// }

// func euCSLRecords(logger log.Logger, initialDir string) ([]csl_eu.CSLRecord, error) {
// 	file, err := csl_eu.DownloadEU(logger, initialDir)
// 	if err != nil {
// 		logger.Warn().Logf("skipping EU CSL download: %v", err)
// 		// no error to return because we skip the download
// 		return nil, nil
// 	}

// 	cslRecords, _, err := csl_eu.ParseEU(file["eu_csl.csv"])
// 	if err != nil {
// 		return nil, err
// 	}
// 	return cslRecords, err

// }

// func ukCSLRecords(logger log.Logger, initialDir string) ([]csl_uk.CSLRecord, error) {
// 	file, err := csl_uk.DownloadCSL(logger, initialDir)
// 	if err != nil {
// 		logger.Warn().Logf("skipping UK CSL download: %v", err)
// 		// no error to return because we skip the download
// 		return nil, nil
// 	}
// 	cslRecords, _, err := csl_uk.ReadCSLFile(file["ConList.csv"])
// 	if err != nil {
// 		return nil, err
// 	}
// 	return cslRecords, err
// }

// func ukSanctionsListRecords(logger log.Logger, initialDir string) ([]csl_uk.SanctionsListRecord, error) {
// 	file, err := csl_uk.DownloadSanctionsList(logger, initialDir)
// 	if file == nil || err != nil {
// 		logger.Warn().Logf("skipping UK Sanctions List download: %v", err)
// 		// no error to return because we skip the download
// 		return nil, nil
// 	}

// 	records, _, err := csl_uk.ReadSanctionsListFile(file["UK_Sanctions_List.ods"])
// 	if err != nil {
// 		return nil, err
// 	}
// 	return records, err
// }

// 	var euCSLs []Result[csl_eu.CSLRecord]
// 	withEUScreeningList := cmp.Or(os.Getenv("WITH_EU_SCREENING_LIST"), "true")
// 	if strx.Yes(withEUScreeningList) {
// 		euConsolidatedList, err := euCSLRecords(s.logger, initialDir)
// 		if err != nil {
// 			stats.Errors = append(stats.Errors, fmt.Errorf("EUCSL: %v", err))
// 		}
// 		euCSLs = precomputeCSLEntities[csl_eu.CSLRecord](euConsolidatedList, s.pipe)
// 	}

// 	var ukCSLs []Result[csl_uk.CSLRecord]
// 	withUKCSLSanctionsList := cmp.Or(os.Getenv("WITH_UK_CSL_SANCTIONS_LIST"), "true")
// 	if strx.Yes(withUKCSLSanctionsList) {
// 		ukConsolidatedList, err := ukCSLRecords(s.logger, initialDir)
// 		if err != nil {
// 			stats.Errors = append(stats.Errors, fmt.Errorf("UKCSL: %v", err))
// 		}
// 		ukCSLs = precomputeCSLEntities[csl_uk.CSLRecord](ukConsolidatedList, s.pipe)
// 	}

// 	var ukSLs []Result[csl_uk.SanctionsListRecord]
// 	withUKSanctionsList := os.Getenv("WITH_UK_SANCTIONS_LIST")
// 	if strings.ToLower(withUKSanctionsList) == "true" {
// 		ukSanctionsList, err := ukSanctionsListRecords(s.logger, initialDir)
// 		if err != nil {
// 			stats.Errors = append(stats.Errors, fmt.Errorf("UKSanctionsList: %v", err))
// 		}
// 		ukSLs = precomputeCSLEntities[csl_uk.SanctionsListRecord](ukSanctionsList, s.pipe)

// 		stats.UKSanctionsList = len(ukSLs)
// 	}

// 	// csl records from US downloaded here
// 	var usConsolidatedLists csl_us.CSL
// 	withUSConsolidatedLists := cmp.Or(os.Getenv("WITH_US_CSL_SANCTIONS_LIST"), "true")
// 	if strx.Yes(withUSConsolidatedLists) {
// 		usConsolidatedLists, err = cslUSRecords(s.logger, initialDir)
// 		if err != nil {
// 			stats.Errors = append(stats.Errors, fmt.Errorf("US CSL: %v", err))
// 		}
// 	}
