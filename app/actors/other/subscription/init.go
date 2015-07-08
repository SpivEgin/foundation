package subscription

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
)

// init makes package self-initialization routine before app start
func init() {
	db.RegisterOnDatabaseStart(setupDB)
	api.RegisterOnRestServiceStart(setupAPI)
	app.OnAppStart(onAppStart)
}

// setupDB prepares system database for package usage
func setupDB() error {

	collection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	collection.AddColumn("order_id", db.ConstTypeID, true)
	collection.AddColumn("visitor_id", db.ConstTypeID, true)

	// a date on which client set a date to bill order
	collection.AddColumn("date", db.ConstTypeDatetime, true)

	// TODO: check database type for duration (we need month or some think like this)
	collection.AddColumn("period", db.ConstTypeDatetime, false)

	collection.AddColumn("status", db.TypeWPrecision(db.ConstTypeVarchar, 50), false)

	return nil
}

// onAppStart makes module initialization on application startup
func onAppStart() error {

	//	if scheduler := env.GetScheduler(); scheduler != nil {
	//		scheduler.RegisterTask("checkOrdersToSent", schedulerFunc)
	//		scheduler.ScheduleRepeat("0 0 * * *", "checkOrdersToSent", nil)
	//	}

	return nil
}
