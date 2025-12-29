use anyhow::{Context, Result};
use directories::ProjectDirs;
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;

const APP_NAME: &str = "todo-cli";
const ORG_NAME: &str = "go-api-starter";

#[derive(Debug, Default, Serialize, Deserialize)]
pub struct Config {
    #[serde(default)]
    pub api_url: Option<String>,

    #[serde(skip)]
    token: Option<String>,

    #[serde(skip)]
    config_path: Option<PathBuf>,
}

impl Config {
    pub fn load() -> Result<Self> {
        let config_path = Self::config_path()?;

        let mut config = if config_path.exists() {
            let content = fs::read_to_string(&config_path)
                .context("Failed to read config file")?;
            toml::from_str(&content).unwrap_or_default()
        } else {
            Config::default()
        };

        config.config_path = Some(config_path);

        // Try to load token from keyring
        config.token = Self::load_token_from_keyring().ok();

        Ok(config)
    }

    pub fn save(&self) -> Result<()> {
        if let Some(ref path) = self.config_path {
            if let Some(parent) = path.parent() {
                fs::create_dir_all(parent).context("Failed to create config directory")?;
            }

            let content = toml::to_string_pretty(self)?;
            fs::write(path, content).context("Failed to write config file")?;
        }
        Ok(())
    }

    fn config_path() -> Result<PathBuf> {
        let proj_dirs = ProjectDirs::from("", ORG_NAME, APP_NAME)
            .context("Failed to determine config directory")?;

        Ok(proj_dirs.config_dir().join("config.toml"))
    }

    pub fn get_token(&self) -> Option<String> {
        self.token.clone()
    }

    pub fn set_token(token: &str) -> Result<()> {
        let entry = keyring::Entry::new(APP_NAME, "api_token")
            .context("Failed to create keyring entry")?;
        entry.set_password(token)
            .context("Failed to save token to keyring")?;
        Ok(())
    }

    pub fn clear_token() -> Result<()> {
        let entry = keyring::Entry::new(APP_NAME, "api_token")
            .context("Failed to create keyring entry")?;
        // Ignore error if token doesn't exist
        let _ = entry.delete_credential();
        Ok(())
    }

    fn load_token_from_keyring() -> Result<String> {
        let entry = keyring::Entry::new(APP_NAME, "api_token")
            .context("Failed to create keyring entry")?;
        entry.get_password()
            .context("Failed to get token from keyring")
    }

    pub fn set_url(&mut self, url: &str) -> Result<()> {
        self.api_url = Some(url.to_string());
        self.save()
    }

    pub fn print(&self) {
        println!("Configuration:");
        println!("  Config file: {:?}", self.config_path);
        println!("  API URL: {}", self.api_url.as_deref().unwrap_or("(default)"));
        println!("  Token: {}", if self.token.is_some() { "✓ stored" } else { "✗ not set" });
    }

    pub fn has_token(&self) -> bool {
        self.token.is_some()
    }
}
