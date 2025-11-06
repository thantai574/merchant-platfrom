package firebase

import "testing"

func Test_repoImpl_SendByToken(t *testing.T) {
	data := make(map[string]interface{})
	data["lixi"] = struct {
		Type string `json:"type"`
	}{
		Type: "test",
	}
	data["type"] = "test"
	type args struct {
		tokens []string
		title  string
		body   string
		data   map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"my device",
			args{
				tokens: []string{
					"c3FgMkQ_Qy6f_yB4Q4brcG:APA91bGxreZETUUgW4QRD8Im_ZNP2WtL-sGVBPhyQg7ix_b_LcHNNGG5qddt2tzPUlFLLVr_Vi7K9SdzyrkMZHxLU8OiHVOq12BM2Y5AnNWbawCvJMwVrhp97-HECRZwXdcYVeVopBvF",
				},
				title: "test notification",
				body:  "test content",
				data:  data,
			},
			false,
		},
	}

	firebaseService := repoImpl{
		FirebaseKey: "AAAASloRZog:APA91bErKucX58X9FHYZD_f8jdwOn-ZoKc3T9AD8WGxOfxdVakiMcAQ1eVmiIKe6Mr_Apqy9PC5-ADbFSWynYQ0T3CDX7u-gy5AbqUH6YC-hkCWD8KDI8dvhJCIyoo6VMTxzFfAX4Oj5",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := firebaseService.SendByToken(tt.args.tokens, tt.args.title, tt.args.body, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("repoImpl.SendByToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
