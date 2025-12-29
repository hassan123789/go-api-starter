use anyhow::Result;
use colored::Colorize;

use crate::api::ApiClient;
use crate::config::Config;

pub async fn login(client: &ApiClient, email: &str, password: &str) -> Result<()> {
    println!("🔑 Logging in as {}...", email);

    let response = client.login(email, password).await?;

    Config::set_token(&response.token)?;

    println!("{}", "✅ Login successful!".green());
    println!("Token has been securely stored.");

    Ok(())
}

pub async fn register(client: &ApiClient, email: &str, password: &str) -> Result<()> {
    println!("📝 Registering {}...", email);

    let response = client.register(email, password).await?;

    Config::set_token(&response.token)?;

    println!("{}", "✅ Registration successful!".green());
    println!("You are now logged in.");

    Ok(())
}

pub fn logout() -> Result<()> {
    Config::clear_token()?;
    println!("{}", "✅ Logged out successfully!".green());
    Ok(())
}

pub fn status(config: &Config) -> Result<()> {
    if config.has_token() {
        println!("{}", "✅ Authenticated".green());
        println!("You are logged in and can access the API.");
    } else {
        println!("{}", "❌ Not authenticated".red());
        println!("Run 'todo auth login' to authenticate.");
    }
    Ok(())
}
