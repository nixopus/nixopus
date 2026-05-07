package llm

import "testing"

func TestLooksLikeUnfinishedPlan(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect bool
	}{
		{"empty", "", false},
		{"normal reply", "The deployment is complete. Your app is live at https://example.com", false},
		{"ill deploy", "I've analyzed the repo. I'll now deploy this as a new application.", true},
		{"let me create", "Repository analysis complete. Let me create the project with the detected settings.", true},
		{"ill proceed", "Good, I can see the repo. I'll proceed with the deployment.", true},
		{"proceeding to", "Everything looks good. Proceeding to create the application.", true},
		{"im going to", "The analysis shows Node.js. I'm going to set up the deployment now.", true},
		{"past tense safe", "I deployed the application successfully.", false},
		{"question safe", "Would you like me to deploy this?", false},
		{"ill now configure", "I'll now configure the environment variables.", true},
		{"let me now check", "Let me now check the deployment status.", true},
		{"now ill", "Now I'll create a feature branch for this fix.", true},
		{"i will now", "I will now send the notification.", true},
		{"plan in middle not at tail", "I'll deploy it soon.\n\nBut first, here is the summary of the repository analysis: the framework is Next.js, port is 3000, and it uses npm as the package manager. The repository has a Dockerfile present. The build configuration looks standard with no monorepo setup detected. Environment variables include DATABASE_URL and REDIS_URL which you will need to configure. The application runs on port 3000 by default with a health check endpoint at /api/health.", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := looksLikeUnfinishedPlan(tc.input)
			if got != tc.expect {
				t.Errorf("looksLikeUnfinishedPlan(%q) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}
