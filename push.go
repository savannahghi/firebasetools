package firebasetools

// FirebaseWebpushConfigInput is used to set the Firebase web config
type FirebaseWebpushConfigInput struct {
	Headers map[string]interface{} `json:"headers"`
	Data    map[string]interface{} `json:"data"`
}

// FirebaseAPNSConfigInput is used to set Apple APNS settings
type FirebaseAPNSConfigInput struct {
	Headers map[string]interface{} `json:"headers"`
}

// FirebaseSimpleNotificationInput is used to create/send simple FCM notifications
type FirebaseSimpleNotificationInput struct {
	Title    string                 `json:"title,omitempty"`
	Body     string                 `json:"body,omitempty"`
	ImageURL *string                `json:"image,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// FirebaseAndroidConfigInput is used to send Firebase Android config values
type FirebaseAndroidConfigInput struct {
	Priority              string                 `json:"priority"` // one of "normal" or "high"
	CollapseKey           *string                `json:"collapseKey"`
	RestrictedPackageName *string                `json:"restrictedPackageName"`
	Data                  map[string]interface{} `json:"data"` // if specified, overrides the Data field on Message type
}

// SendNotificationPayload is used to serialise and save incoming Send FCM notification requests.
type SendNotificationPayload struct {
	RegistrationTokens []string                         `json:"registrationTokens"`
	Data               map[string]string                `json:"data"`
	Notification       *FirebaseSimpleNotificationInput `json:"notification"`
	Android            *FirebaseAndroidConfigInput      `json:"android"`
	Ios                *FirebaseAPNSConfigInput         `json:"ios"`
	Web                *FirebaseWebpushConfigInput      `json:"web"`
}
