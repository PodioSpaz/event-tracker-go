package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewAttendee(t *testing.T) {
	activityID := int64(1)
	personID := int64(2)
	role := RoleParticipant

	attendee := NewAttendee(activityID, personID, role)

	assert.NotNil(t, attendee)
	assert.Equal(t, activityID, attendee.ActivityID)
	assert.Equal(t, personID, attendee.PersonID)
	assert.Equal(t, role, attendee.Role)
	assert.Equal(t, PaymentStatusUnpaid, attendee.PaymentStatus)
	assert.True(t, attendee.PaymentAmount.IsZero())
}

func TestAttendee_IsPaid(t *testing.T) {
	attendee := &Attendee{PaymentStatus: PaymentStatusPaid}
	assert.True(t, attendee.IsPaid())

	attendee.PaymentStatus = PaymentStatusWaived
	assert.True(t, attendee.IsPaid())

	attendee.PaymentStatus = PaymentStatusUnpaid
	assert.False(t, attendee.IsPaid())
}

func TestAttendee_IsUnpaid(t *testing.T) {
	attendee := &Attendee{PaymentStatus: PaymentStatusUnpaid}
	assert.True(t, attendee.IsUnpaid())

	attendee.PaymentStatus = PaymentStatusPaid
	assert.False(t, attendee.IsUnpaid())
}

func TestAttendee_IsWaived(t *testing.T) {
	attendee := &Attendee{PaymentStatus: PaymentStatusWaived}
	assert.True(t, attendee.IsWaived())

	attendee.PaymentStatus = PaymentStatusPaid
	assert.False(t, attendee.IsWaived())
}

func TestAttendee_MarkPaid(t *testing.T) {
	attendee := &Attendee{}
	amount := decimal.NewFromFloat(25.50)
	date := time.Now()

	err := attendee.MarkPaid(amount, date)
	assert.NoError(t, err)
	assert.Equal(t, PaymentStatusPaid, attendee.PaymentStatus)
	assert.Equal(t, amount, attendee.PaymentAmount)
	assert.NotNil(t, attendee.PaymentDate)
	assert.Equal(t, date, *attendee.PaymentDate)

	// Negative amount
	err = attendee.MarkPaid(decimal.NewFromFloat(-10), date)
	assert.Error(t, err)
}

func TestAttendee_MarkPaidNow(t *testing.T) {
	attendee := &Attendee{}
	amount := decimal.NewFromFloat(25.50)

	err := attendee.MarkPaidNow(amount)
	assert.NoError(t, err)
	assert.Equal(t, PaymentStatusPaid, attendee.PaymentStatus)
	assert.Equal(t, amount, attendee.PaymentAmount)
	assert.NotNil(t, attendee.PaymentDate)
}

func TestAttendee_MarkUnpaid(t *testing.T) {
	date := time.Now()
	attendee := &Attendee{
		PaymentStatus: PaymentStatusPaid,
		PaymentAmount: decimal.NewFromFloat(25.50),
		PaymentDate:   &date,
	}

	attendee.MarkUnpaid()
	assert.Equal(t, PaymentStatusUnpaid, attendee.PaymentStatus)
	assert.True(t, attendee.PaymentAmount.IsZero())
	assert.Nil(t, attendee.PaymentDate)
}

func TestAttendee_WaivePayment(t *testing.T) {
	attendee := &Attendee{}
	date := time.Now()

	attendee.WaivePayment(date)
	assert.Equal(t, PaymentStatusWaived, attendee.PaymentStatus)
	assert.True(t, attendee.PaymentAmount.IsZero())
	assert.NotNil(t, attendee.PaymentDate)
	assert.Equal(t, date, *attendee.PaymentDate)
}

func TestAttendee_WaivePaymentNow(t *testing.T) {
	attendee := &Attendee{}

	attendee.WaivePaymentNow()
	assert.Equal(t, PaymentStatusWaived, attendee.PaymentStatus)
	assert.True(t, attendee.PaymentAmount.IsZero())
	assert.NotNil(t, attendee.PaymentDate)
}

func TestAttendee_SetPaymentAmount(t *testing.T) {
	attendee := &Attendee{}

	// Valid amount
	err := attendee.SetPaymentAmount(decimal.NewFromFloat(25.50))
	assert.NoError(t, err)
	assert.Equal(t, "25.5", attendee.PaymentAmount.String())

	// Negative amount
	err = attendee.SetPaymentAmount(decimal.NewFromFloat(-10))
	assert.Error(t, err)
}

func TestAttendee_SetRole(t *testing.T) {
	attendee := &Attendee{Role: RoleParticipant}

	// Valid role
	err := attendee.SetRole(RoleVolunteer)
	assert.NoError(t, err)
	assert.Equal(t, RoleVolunteer, attendee.Role)

	// Invalid role
	err = attendee.SetRole(AttendeeRole("invalid"))
	assert.Error(t, err)
}

func TestAttendee_GetRoleDisplay(t *testing.T) {
	tests := []struct {
		role     AttendeeRole
		expected string
	}{
		{RoleParticipant, "Participant"},
		{RoleVolunteer, "Volunteer"},
		{RoleWorshipTeam, "Worship Team"},
		{RoleWorkshopLeader, "Workshop Leader"},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			attendee := &Attendee{Role: tt.role}
			assert.Equal(t, tt.expected, attendee.GetRoleDisplay())
		})
	}
}

func TestAttendee_DaysSinceRegistration(t *testing.T) {
	attendee := &Attendee{
		RegistrationDate: time.Now().AddDate(0, 0, -5),
	}

	days := attendee.DaysSinceRegistration()
	assert.Equal(t, 5, days)
}

func TestAttendee_DaysSincePayment(t *testing.T) {
	// No payment date
	attendee := &Attendee{}
	assert.Equal(t, 0, attendee.DaysSincePayment())

	// With payment date
	paymentDate := time.Now().AddDate(0, 0, -3)
	attendee.PaymentDate = &paymentDate
	days := attendee.DaysSincePayment()
	assert.Equal(t, 3, days)
}

func TestAttendee_Validate(t *testing.T) {
	validAttendee := &Attendee{
		ActivityID:    1,
		PersonID:      2,
		Role:          RoleParticipant,
		PaymentStatus: PaymentStatusUnpaid,
		PaymentAmount: decimal.Zero,
	}

	err := validAttendee.Validate()
	assert.NoError(t, err)

	// Missing activity ID
	invalidAttendee := *validAttendee
	invalidAttendee.ActivityID = 0
	assert.Error(t, invalidAttendee.Validate())

	// Missing person ID
	invalidAttendee = *validAttendee
	invalidAttendee.PersonID = 0
	assert.Error(t, invalidAttendee.Validate())

	// Invalid role
	invalidAttendee = *validAttendee
	invalidAttendee.Role = AttendeeRole("invalid")
	assert.Error(t, invalidAttendee.Validate())

	// Invalid payment status
	invalidAttendee = *validAttendee
	invalidAttendee.PaymentStatus = PaymentStatus("invalid")
	assert.Error(t, invalidAttendee.Validate())

	// Negative payment amount
	invalidAttendee = *validAttendee
	invalidAttendee.PaymentAmount = decimal.NewFromFloat(-10)
	assert.Error(t, invalidAttendee.Validate())

	// Paid with no amount
	invalidAttendee = *validAttendee
	invalidAttendee.PaymentStatus = PaymentStatusPaid
	invalidAttendee.PaymentAmount = decimal.Zero
	assert.Error(t, invalidAttendee.Validate())

	// Paid with no date
	invalidAttendee = *validAttendee
	invalidAttendee.PaymentStatus = PaymentStatusPaid
	invalidAttendee.PaymentAmount = decimal.NewFromFloat(25.50)
	invalidAttendee.PaymentDate = nil
	assert.Error(t, invalidAttendee.Validate())

	// Waived with amount
	invalidAttendee = *validAttendee
	invalidAttendee.PaymentStatus = PaymentStatusWaived
	invalidAttendee.PaymentAmount = decimal.NewFromFloat(25.50)
	assert.Error(t, invalidAttendee.Validate())
}

func TestAttendee_ValidatePaymentAmount(t *testing.T) {
	activityFee := decimal.NewFromFloat(25.50)

	// Matching amount
	attendee := &Attendee{
		PaymentStatus: PaymentStatusPaid,
		PaymentAmount: activityFee,
	}
	assert.NoError(t, attendee.ValidatePaymentAmount(activityFee))

	// Mismatched amount
	attendee.PaymentAmount = decimal.NewFromFloat(10.00)
	assert.Error(t, attendee.ValidatePaymentAmount(activityFee))

	// Waived (doesn't need to match)
	attendee.PaymentStatus = PaymentStatusWaived
	attendee.PaymentAmount = decimal.Zero
	assert.NoError(t, attendee.ValidatePaymentAmount(activityFee))

	// Unpaid
	attendee.PaymentStatus = PaymentStatusUnpaid
	assert.NoError(t, attendee.ValidatePaymentAmount(activityFee))
}
