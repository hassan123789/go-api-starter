use anyhow::{Context, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Todo {
    pub id: i64,
    pub user_id: i64,
    pub title: String,
    pub completed: bool,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct TodoListResponse {
    pub todos: Vec<Todo>,
    pub total: i32,
}

#[derive(Debug, Serialize)]
struct CreateTodoRequest {
    title: String,
}

#[derive(Debug, Serialize)]
struct UpdateTodoRequest {
    #[serde(skip_serializing_if = "Option::is_none")]
    title: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    completed: Option<bool>,
}

#[derive(Debug, Serialize)]
pub struct LoginRequest {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Serialize)]
pub struct RegisterRequest {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct AuthResponse {
    pub token: String,
    #[serde(default)]
    pub user_id: Option<i64>,
}

#[derive(Debug, Deserialize)]
struct ApiError {
    error: String,
}

pub struct ApiClient {
    client: Client,
    base_url: String,
    token: Option<String>,
}

impl ApiClient {
    pub fn new(base_url: &str, token: Option<String>) -> Self {
        let client = Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .build()
            .expect("Failed to create HTTP client");

        Self {
            client,
            base_url: base_url.trim_end_matches('/').to_string(),
            token,
        }
    }

    #[allow(dead_code)]
    pub fn with_token(&self, token: String) -> Self {
        Self {
            client: self.client.clone(),
            base_url: self.base_url.clone(),
            token: Some(token),
        }
    }

    fn auth_header(&self) -> Option<String> {
        self.token.as_ref().map(|t| format!("Bearer {}", t))
    }

    pub async fn login(&self, email: &str, password: &str) -> Result<AuthResponse> {
        let url = format!("{}/api/v1/auth/login", self.base_url);

        let response = self
            .client
            .post(&url)
            .json(&LoginRequest {
                email: email.to_string(),
                password: password.to_string(),
            })
            .send()
            .await
            .context("Failed to send login request")?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await.unwrap_or(ApiError {
                error: "Unknown error".to_string(),
            });
            anyhow::bail!("Login failed: {}", error.error);
        }

        response.json().await.context("Failed to parse login response")
    }

    pub async fn register(&self, email: &str, password: &str) -> Result<AuthResponse> {
        let url = format!("{}/api/v1/users", self.base_url);

        let response = self
            .client
            .post(&url)
            .json(&RegisterRequest {
                email: email.to_string(),
                password: password.to_string(),
            })
            .send()
            .await
            .context("Failed to send register request")?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await.unwrap_or(ApiError {
                error: "Unknown error".to_string(),
            });
            anyhow::bail!("Registration failed: {}", error.error);
        }

        response.json().await.context("Failed to parse register response")
    }

    pub async fn list_todos(&self, _completed: Option<bool>) -> Result<Vec<Todo>> {
        let url = format!("{}/api/v1/todos", self.base_url);

        let mut request = self.client.get(&url);
        if let Some(auth) = self.auth_header() {
            request = request.header("Authorization", auth);
        }

        let response = request.send().await.context("Failed to fetch todos")?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await.unwrap_or(ApiError {
                error: "Unknown error".to_string(),
            });
            anyhow::bail!("Failed to list todos: {}", error.error);
        }

        let list: TodoListResponse = response.json().await.context("Failed to parse todos")?;
        Ok(list.todos)
    }

    pub async fn get_todo(&self, id: i64) -> Result<Todo> {
        let url = format!("{}/api/v1/todos/{}", self.base_url, id);

        let mut request = self.client.get(&url);
        if let Some(auth) = self.auth_header() {
            request = request.header("Authorization", auth);
        }

        let response = request.send().await.context("Failed to fetch todo")?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await.unwrap_or(ApiError {
                error: "Unknown error".to_string(),
            });
            anyhow::bail!("Failed to get todo: {}", error.error);
        }

        response.json().await.context("Failed to parse todo")
    }

    pub async fn create_todo(&self, title: &str) -> Result<Todo> {
        let url = format!("{}/api/v1/todos", self.base_url);

        let mut request = self.client.post(&url).json(&CreateTodoRequest {
            title: title.to_string(),
        });
        if let Some(auth) = self.auth_header() {
            request = request.header("Authorization", auth);
        }

        let response = request.send().await.context("Failed to create todo")?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await.unwrap_or(ApiError {
                error: "Unknown error".to_string(),
            });
            anyhow::bail!("Failed to create todo: {}", error.error);
        }

        response.json().await.context("Failed to parse created todo")
    }

    pub async fn update_todo(
        &self,
        id: i64,
        title: Option<&str>,
        completed: Option<bool>,
    ) -> Result<Todo> {
        let url = format!("{}/api/v1/todos/{}", self.base_url, id);

        let mut request = self.client.put(&url).json(&UpdateTodoRequest {
            title: title.map(|s| s.to_string()),
            completed,
        });
        if let Some(auth) = self.auth_header() {
            request = request.header("Authorization", auth);
        }

        let response = request.send().await.context("Failed to update todo")?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await.unwrap_or(ApiError {
                error: "Unknown error".to_string(),
            });
            anyhow::bail!("Failed to update todo: {}", error.error);
        }

        response.json().await.context("Failed to parse updated todo")
    }

    pub async fn delete_todo(&self, id: i64) -> Result<()> {
        let url = format!("{}/api/v1/todos/{}", self.base_url, id);

        let mut request = self.client.delete(&url);
        if let Some(auth) = self.auth_header() {
            request = request.header("Authorization", auth);
        }

        let response = request.send().await.context("Failed to delete todo")?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await.unwrap_or(ApiError {
                error: "Unknown error".to_string(),
            });
            anyhow::bail!("Failed to delete todo: {}", error.error);
        }

        Ok(())
    }
}
