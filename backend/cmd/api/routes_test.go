package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMessageHistoryRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/messages/1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestConversationReadsRequireAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	for _, path := range []string{"/conversation/clerk-user/1", "/conversations/clerk-user"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("path %s: status = %d, want %d", path, recorder.Code, http.StatusForbidden)
		}
	}
}

func TestFriendReadsRequireAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	for _, path := range []string{"/friends/clerk-user", "/requests/clerk-user"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("path %s: status = %d, want %d", path, recorder.Code, http.StatusForbidden)
		}
	}
}

func TestCreateFriendRequestRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodPost, "/request", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestAcceptFriendRequestRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodPost, "/request/accept/1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestDenyFriendRequestRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodDelete, "/deny/1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestDeleteFriendRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodDelete, "/friend/1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestLeaveGroupRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodDelete, "/group/leave/attacker/1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestUnseenMessageCountRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/message/unseen/clerk-user", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestLastMessageRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/message/1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestCreateMessageRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodPost, "/message", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestMarkMessageReadRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodPost, "/message/mark_read", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestHealthRemainsPublic(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestReadinessFailsWithoutDatabase(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
