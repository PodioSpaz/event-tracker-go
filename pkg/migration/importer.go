package migration

import (
	"context"
	"fmt"

	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/rs/zerolog/log"
)

// ImportStats contains statistics about the import operation
type ImportStats struct {
	ActivitiesImported int
	PeopleImported     int
	AttendeesImported  int
	ActivitiesFailed   int
	PeopleFailed       int
	AttendeesFailed    int
	Errors             []string
}

// Importer handles importing TinyDB data into the database
type Importer struct {
	parser       *TinyDBParser
	mapper       *Mapper
	activityRepo repository.ActivityRepository
	personRepo   repository.PersonRepository
	attendeeRepo repository.AttendeeRepository
}

// NewImporter creates a new importer
func NewImporter(
	activityRepo repository.ActivityRepository,
	personRepo repository.PersonRepository,
	attendeeRepo repository.AttendeeRepository,
) *Importer {
	return &Importer{
		parser:       NewTinyDBParser(),
		mapper:       NewMapper(),
		activityRepo: activityRepo,
		personRepo:   personRepo,
		attendeeRepo: attendeeRepo,
	}
}

// ImportFile imports data from a TinyDB JSON file
func (i *Importer) ImportFile(ctx context.Context, filePath string) (*ImportStats, error) {
	log.Info().Str("file", filePath).Msg("Starting import from TinyDB file")

	// Parse file
	doc, err := i.parser.ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Validate document
	if err := i.parser.Validate(doc); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Import data
	stats, err := i.ImportDocument(ctx, doc)
	if err != nil {
		return stats, fmt.Errorf("import failed: %w", err)
	}

	log.Info().
		Int("activities", stats.ActivitiesImported).
		Int("people", stats.PeopleImported).
		Int("attendees", stats.AttendeesImported).
		Msg("Import completed successfully")

	return stats, nil
}

// ImportDocument imports a parsed TinyDB document
func (i *Importer) ImportDocument(ctx context.Context, doc *TinyDBDocument) (*ImportStats, error) {
	stats := &ImportStats{
		Errors: make([]string, 0),
	}

	// Import people first (no dependencies)
	personIDMap, err := i.importPeople(ctx, doc, stats)
	if err != nil {
		return stats, fmt.Errorf("failed to import people: %w", err)
	}

	// Import activities (no dependencies)
	activityIDMap, err := i.importActivities(ctx, doc, stats)
	if err != nil {
		return stats, fmt.Errorf("failed to import activities: %w", err)
	}

	// Import attendees (depends on people and activities)
	if err := i.importAttendees(ctx, doc, activityIDMap, personIDMap, stats); err != nil {
		return stats, fmt.Errorf("failed to import attendees: %w", err)
	}

	return stats, nil
}

// importPeople imports all people and returns a map of TinyDB ID to SQLite ID
func (i *Importer) importPeople(ctx context.Context, doc *TinyDBDocument, stats *ImportStats) (map[int]int64, error) {
	log.Info().Int("count", len(doc.People)).Msg("Importing people")

	personIDMap := make(map[int]int64)

	for id, tdbPerson := range doc.People {
		// Map to domain model
		person, err := i.mapper.MapPerson(&tdbPerson)
		if err != nil {
			stats.PeopleFailed++
			errMsg := fmt.Sprintf("Failed to map person %s: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Failed to map person")
			continue
		}

		// Validate
		if err := person.Validate(); err != nil {
			stats.PeopleFailed++
			errMsg := fmt.Sprintf("Person %s validation failed: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Person validation failed")
			continue
		}

		// Check for duplicate email
		if person.Email != "" {
			exists, err := i.personRepo.ExistsByEmail(ctx, person.Email)
			if err != nil {
				stats.PeopleFailed++
				errMsg := fmt.Sprintf("Failed to check email for person %s: %v", id, err)
				stats.Errors = append(stats.Errors, errMsg)
				log.Warn().Str("id", id).Err(err).Msg("Failed to check email")
				continue
			}
			if exists {
				// Try to find existing person
				existing, err := i.personRepo.FindByEmail(ctx, person.Email)
				if err == nil && existing != nil {
					personIDMap[tdbPerson.DocID] = existing.ID
					stats.PeopleImported++
					log.Info().Str("email", person.Email).Msg("Person already exists, using existing ID")
					continue
				}
			}
		}

		// Create person
		if err := i.personRepo.Create(ctx, person); err != nil {
			stats.PeopleFailed++
			errMsg := fmt.Sprintf("Failed to create person %s: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Failed to create person")
			continue
		}

		// Store ID mapping
		personIDMap[tdbPerson.DocID] = person.ID
		stats.PeopleImported++

		log.Debug().
			Str("id", id).
			Int64("newID", person.ID).
			Str("name", person.FullName()).
			Msg("Person imported")
	}

	return personIDMap, nil
}

// importActivities imports all activities and returns a map of TinyDB ID to SQLite ID
func (i *Importer) importActivities(ctx context.Context, doc *TinyDBDocument, stats *ImportStats) (map[int]int64, error) {
	log.Info().Int("count", len(doc.Activities)).Msg("Importing activities")

	activityIDMap := make(map[int]int64)

	for id, tdbActivity := range doc.Activities {
		// Map to domain model
		activity, err := i.mapper.MapActivity(&tdbActivity)
		if err != nil {
			stats.ActivitiesFailed++
			errMsg := fmt.Sprintf("Failed to map activity %s: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Failed to map activity")
			continue
		}

		// Validate
		if err := activity.Validate(); err != nil {
			stats.ActivitiesFailed++
			errMsg := fmt.Sprintf("Activity %s validation failed: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Activity validation failed")
			continue
		}

		// Create activity
		if err := i.activityRepo.Create(ctx, activity); err != nil {
			stats.ActivitiesFailed++
			errMsg := fmt.Sprintf("Failed to create activity %s: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Failed to create activity")
			continue
		}

		// Store ID mapping
		activityIDMap[tdbActivity.DocID] = activity.ID
		stats.ActivitiesImported++

		log.Debug().
			Str("id", id).
			Int64("newID", activity.ID).
			Str("name", activity.Name).
			Msg("Activity imported")
	}

	return activityIDMap, nil
}

// importAttendees imports all attendees
func (i *Importer) importAttendees(
	ctx context.Context,
	doc *TinyDBDocument,
	activityIDMap, personIDMap map[int]int64,
	stats *ImportStats,
) error {
	log.Info().Int("count", len(doc.Attendees)).Msg("Importing attendees")

	for id, tdbAttendee := range doc.Attendees {
		// Map to domain model
		attendee, err := i.mapper.MapAttendee(&tdbAttendee, activityIDMap, personIDMap)
		if err != nil {
			stats.AttendeesFailed++
			errMsg := fmt.Sprintf("Failed to map attendee %s: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Failed to map attendee")
			continue
		}

		// Validate
		if err := attendee.Validate(); err != nil {
			stats.AttendeesFailed++
			errMsg := fmt.Sprintf("Attendee %s validation failed: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Attendee validation failed")
			continue
		}

		// Check for duplicate registration
		isRegistered, err := i.attendeeRepo.IsRegistered(ctx, attendee.ActivityID, attendee.PersonID)
		if err != nil {
			stats.AttendeesFailed++
			errMsg := fmt.Sprintf("Failed to check registration for attendee %s: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Failed to check registration")
			continue
		}
		if isRegistered {
			log.Info().
				Str("id", id).
				Int64("activityID", attendee.ActivityID).
				Int64("personID", attendee.PersonID).
				Msg("Attendee already registered, skipping")
			stats.AttendeesImported++
			continue
		}

		// Create attendee
		if err := i.attendeeRepo.Create(ctx, attendee); err != nil {
			stats.AttendeesFailed++
			errMsg := fmt.Sprintf("Failed to create attendee %s: %v", id, err)
			stats.Errors = append(stats.Errors, errMsg)
			log.Warn().Str("id", id).Err(err).Msg("Failed to create attendee")
			continue
		}

		stats.AttendeesImported++

		log.Debug().
			Str("id", id).
			Int64("newID", attendee.ID).
			Msg("Attendee imported")
	}

	return nil
}
