package ui

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"github.com/turtacn/QuantaID/web"

	"github.com/turtacn/QuantaID/internal/server/http/middleware"
)

// TestLoginRenderIsolation focuses ONLY on rendering the login page
// to isolate template parsing and inheritance issues.
func TestLoginRenderIsolation(t *testing.T) {
	// Step 1: Manually parse ONLY the layout and login templates.
	// This avoids parsing other potentially broken templates.
	templates, err := template.ParseFS(web.TemplateFS, "templates/*.html")
	require.NoError(t, err, "Failed to parse layout.html and login.html")

	// Step 2: Create a renderer with this limited template set.
	renderer := &Renderer{templates: templates}

	// Step 3: Serve the request to the login page handler.
	req := httptest.NewRequest("GET", "/auth/login", nil)
	rr := httptest.NewRecorder()
	renderer.Render(rr, req, "login.html", nil)

	// Step 4: Assert that the response is correct.
	assert.Equal(t, http.StatusOK, rr.Code, "Expected a 200 OK status for the login page")

	// Step 5: Use goquery to verify the HTML content.
	for _, tmpl := range renderer.templates.Templates() {
		t.Log(tmpl.Name())
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rr.Body.String()))
	require.NoError(t, err, "Failed to parse the rendered HTML")

	// Check that the form and a key input field exist, proving the content block was rendered.
	usernameInput := doc.Find("input[name='username']")
	assert.Equal(t, 1, usernameInput.Length(), "Expected to find exactly one username input field")

	h2 := doc.Find("h2")
	assert.Contains(t, h2.Text(), "Login", "Expected to find the 'Login' header")
}

// setupTestServer configures a test server with the UI handlers and middleware.
func setupTestServer() (http.Handler, error) {
	renderer, err := NewRenderer()
	if err != nil {
		return nil, err
	}

	authHandler := NewAuthHandler(renderer)

	r := mux.NewRouter()
	r.HandleFunc("/auth/login", authHandler.ShowLoginPage).Methods("GET")
	r.HandleFunc("/auth/login", authHandler.HandleLogin).Methods("POST")
	r.HandleFunc("/auth/mfa", authHandler.ShowMFAPage).Methods("GET")

	// Wrap the router with the CSRF middleware for a realistic test
	return middleware.CSRFMiddleware(r), nil
}

func TestAuthHandler_ShowLoginPage(t *testing.T) {
	server, err := setupTestServer()
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/auth/login", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check for the CSRF cookie
	csrfCookie := rr.Result().Cookies()[0]
	assert.Equal(t, "_csrf", csrfCookie.Name)
	assert.NotEmpty(t, csrfCookie.Value)

	// Use goquery to parse the HTML response
	doc, err := goquery.NewDocumentFromReader(rr.Body)
	require.NoError(t, err)

	// Check that the form and CSRF field exist
	usernameInput := doc.Find("input[name='username']")
	assert.Equal(t, 1, usernameInput.Length(), "Expected to find a username input field")

	csrfInput := doc.Find("input[name='_csrf']")
	assert.Equal(t, 1, csrfInput.Length(), "Expected to find a CSRF hidden field")

	csrfToken, exists := csrfInput.Attr("value")
	assert.True(t, exists, "CSRF input should have a value attribute")
	assert.NotEmpty(t, csrfToken, "CSRF token should not be empty")

	// The token in the form should match the one set in the context (which we can't directly check here)
	// but it should be a non-empty string.
}

func TestAuthHandler_HandleLogin(t *testing.T) {
	server, err := setupTestServer()
	require.NoError(t, err)

	// --- Step 1: GET the login page to get a valid CSRF token and cookie ---
	reqGet := httptest.NewRequest("GET", "/auth/login", nil)
	rrGet := httptest.NewRecorder()
	server.ServeHTTP(rrGet, reqGet)
	require.Equal(t, http.StatusOK, rrGet.Code)

	csrfCookie := rrGet.Result().Cookies()[0]
	doc, err := goquery.NewDocumentFromReader(rrGet.Body)
	require.NoError(t, err)
	csrfToken, _ := doc.Find("input[name='_csrf']").Attr("value")

	// --- Test Case 1: Successful Login ---
	t.Run("Successful Login", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "admin")
		form.Add("password", "password")
		form.Add("_csrf", csrfToken)

		reqPost := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
		reqPost.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		reqPost.AddCookie(csrfCookie) // Add the cookie from the GET request
		rrPost := httptest.NewRecorder()

		server.ServeHTTP(rrPost, reqPost)

		// Assert a redirect to the MFA page
		assert.Equal(t, http.StatusFound, rrPost.Code)
		assert.Equal(t, "/auth/mfa", rrPost.Header().Get("Location"))
	})

	// --- Test Case 2: Failed Login (Bad Password) ---
	t.Run("Failed Login", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "admin")
		form.Add("password", "wrongpassword")
		form.Add("_csrf", csrfToken)

		reqPost := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
		reqPost.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		reqPost.AddCookie(csrfCookie)
		rrPost := httptest.NewRecorder()

		server.ServeHTTP(rrPost, reqPost)

		// Assert that the login page is re-rendered with an error
		assert.Equal(t, http.StatusOK, rrPost.Code)
		doc, err := goquery.NewDocumentFromReader(rrPost.Body)
		require.NoError(t, err)
		assert.Contains(t, doc.Find(".error").Text(), "Invalid username or password")
	})

	// --- Test Case 3: Failed Login (CSRF Mismatch) ---
	t.Run("CSRF Mismatch", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "admin")
		form.Add("password", "password")
		form.Add("_csrf", "invalid-csrf-token") // Invalid token

		reqPost := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
		reqPost.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		reqPost.AddCookie(csrfCookie) // Correct cookie, but mismatched form value
		rrPost := httptest.NewRecorder()

		server.ServeHTTP(rrPost, reqPost)

		// Assert Forbidden status
		assert.Equal(t, http.StatusForbidden, rrPost.Code)
	})
}
