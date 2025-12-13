package ui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"github.com/PodioSpaz/event-tracker-go/internal/controller"
	"github.com/PodioSpaz/event-tracker-go/internal/repository/sqlite"
	"github.com/PodioSpaz/event-tracker-go/internal/service"
	"github.com/rs/zerolog/log"
)

// App represents the main application
type App struct {
	fyneApp          fyne.App
	mainWindow       fyne.Window
	db               *sqlite.DB
	activityCtrl     *controller.ActivityController
	personCtrl       *controller.PersonController
	currentView      string
	contentContainer *fyne.Container
	ctx              context.Context
}

// NewApp creates a new application instance
func NewApp(db *sqlite.DB) *App {
	fyneApp := app.NewWithID("com.podiospaz.event-tracker")
	fyneApp.Settings().SetTheme(theme.DefaultTheme())

	// Create repositories
	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)

	// Create services
	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	paymentSvc := service.NewPaymentService(activityRepo, attendeeRepo, db)
	capacitySvc := service.NewCapacityService(activityRepo, attendeeRepo)

	// Create controllers
	activityCtrl := controller.NewActivityController(activityRepo, registrationSvc, paymentSvc, capacitySvc, db)
	personCtrl := controller.NewPersonController(personRepo, registrationSvc, db)

	return &App{
		fyneApp:      fyneApp,
		db:           db,
		activityCtrl: activityCtrl,
		personCtrl:   personCtrl,
		ctx:          context.Background(),
	}
}

// Run starts the application
func (a *App) Run() {
	log.Info().Msg("Starting Event Tracker GUI")

	// Create main window
	a.mainWindow = a.fyneApp.NewWindow("Event Tracker")
	a.mainWindow.Resize(fyne.NewSize(1200, 800))
	a.mainWindow.CenterOnScreen()

	// Create content container
	a.contentContainer = container.NewMax()

	// Create main layout with navigation
	mainContent := a.createMainLayout()
	a.mainWindow.SetContent(mainContent)

	// Show dashboard by default
	a.showDashboard()

	// Handle window close
	a.mainWindow.SetOnClosed(func() {
		log.Info().Msg("Application closing")
		if err := a.db.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing database")
		}
	})

	// Show window and run
	a.mainWindow.ShowAndRun()
}

// createMainLayout creates the main application layout with navigation
func (a *App) createMainLayout() fyne.CanvasObject {
	// Create navigation sidebar
	nav := a.createNavigation()

	// Create split layout with navigation and content
	split := container.NewHSplit(
		nav,
		a.contentContainer,
	)
	split.Offset = 0.15 // Navigation takes 15% of width

	return split
}

// setContent updates the main content area
func (a *App) setContent(content fyne.CanvasObject, viewName string) {
	a.currentView = viewName
	a.contentContainer.Objects = []fyne.CanvasObject{content}
	a.contentContainer.Refresh()
	log.Debug().Str("view", viewName).Msg("Switched to view")
}
