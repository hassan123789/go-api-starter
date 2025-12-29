use anyhow::Result;
use clap::{Parser, Subcommand};

mod api;
mod auth;
mod config;
mod output;

use api::ApiClient;
use config::Config;

/// todo-cli: A CLI tool for managing todos via the go-api-starter API
#[derive(Parser)]
#[command(name = "todo")]
#[command(author, version, about, long_about = None)]
#[command(propagate_version = true)]
struct Cli {
    /// API server URL
    #[arg(short, long, env = "TODO_API_URL", default_value = "http://localhost:8080")]
    url: String,

    /// Output format (text, json)
    #[arg(short, long, default_value = "text")]
    format: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Authentication commands
    Auth {
        #[command(subcommand)]
        command: AuthCommands,
    },
    /// List all todos
    List {
        /// Filter by completion status
        #[arg(short, long)]
        completed: Option<bool>,
    },
    /// Get a specific todo by ID
    Get {
        /// Todo ID
        id: i64,
    },
    /// Create a new todo
    Create {
        /// Todo title
        title: String,
    },
    /// Update a todo
    Update {
        /// Todo ID
        id: i64,
        /// New title
        #[arg(short, long)]
        title: Option<String>,
        /// Mark as completed
        #[arg(short, long)]
        completed: Option<bool>,
    },
    /// Delete a todo
    Delete {
        /// Todo ID
        id: i64,
        /// Skip confirmation
        #[arg(short, long)]
        force: bool,
    },
    /// Mark a todo as completed
    Done {
        /// Todo ID
        id: i64,
    },
    /// Mark a todo as incomplete
    Undone {
        /// Todo ID
        id: i64,
    },
    /// Show configuration
    Config {
        #[command(subcommand)]
        command: Option<ConfigCommands>,
    },
}

#[derive(Subcommand)]
enum AuthCommands {
    /// Login to the API
    Login {
        /// Email address
        #[arg(short, long)]
        email: String,
        /// Password (will prompt if not provided)
        #[arg(short, long)]
        password: Option<String>,
    },
    /// Register a new account
    Register {
        /// Email address
        #[arg(short, long)]
        email: String,
        /// Password (will prompt if not provided)
        #[arg(short, long)]
        password: Option<String>,
    },
    /// Logout (clear stored token)
    Logout,
    /// Show current auth status
    Status,
}

#[derive(Subcommand)]
enum ConfigCommands {
    /// Show current configuration
    Show,
    /// Set API URL
    SetUrl {
        /// API URL
        url: String,
    },
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();
    let config = Config::load()?;
    let client = ApiClient::new(&cli.url, config.get_token());

    match cli.command {
        Commands::Auth { command } => match command {
            AuthCommands::Login { email, password } => {
                let password = password.unwrap_or_else(|| {
                    rpassword_prompt("Password: ")
                });
                auth::login(&client, &email, &password).await?;
            }
            AuthCommands::Register { email, password } => {
                let password = password.unwrap_or_else(|| {
                    rpassword_prompt("Password: ")
                });
                auth::register(&client, &email, &password).await?;
            }
            AuthCommands::Logout => {
                auth::logout()?;
            }
            AuthCommands::Status => {
                auth::status(&config)?;
            }
        },
        Commands::List { completed } => {
            let todos = client.list_todos(completed).await?;
            output::print_todos(&todos, &cli.format)?;
        }
        Commands::Get { id } => {
            let todo = client.get_todo(id).await?;
            output::print_todo(&todo, &cli.format)?;
        }
        Commands::Create { title } => {
            let todo = client.create_todo(&title).await?;
            output::print_todo(&todo, &cli.format)?;
            println!("✅ Todo created successfully!");
        }
        Commands::Update { id, title, completed } => {
            let todo = client.update_todo(id, title.as_deref(), completed).await?;
            output::print_todo(&todo, &cli.format)?;
            println!("✅ Todo updated successfully!");
        }
        Commands::Delete { id, force } => {
            if !force {
                println!("Are you sure you want to delete todo #{}? [y/N]", id);
                let mut input = String::new();
                std::io::stdin().read_line(&mut input)?;
                if !input.trim().eq_ignore_ascii_case("y") {
                    println!("Cancelled.");
                    return Ok(());
                }
            }
            client.delete_todo(id).await?;
            println!("✅ Todo #{} deleted successfully!", id);
        }
        Commands::Done { id } => {
            let todo = client.update_todo(id, None, Some(true)).await?;
            output::print_todo(&todo, &cli.format)?;
            println!("✅ Todo marked as completed!");
        }
        Commands::Undone { id } => {
            let todo = client.update_todo(id, None, Some(false)).await?;
            output::print_todo(&todo, &cli.format)?;
            println!("✅ Todo marked as incomplete!");
        }
        Commands::Config { command } => {
            match command {
                Some(ConfigCommands::Show) | None => {
                    config.print();
                }
                Some(ConfigCommands::SetUrl { url }) => {
                    let mut config = config;
                    config.set_url(&url)?;
                    println!("✅ API URL set to: {}", url);
                }
            }
        }
    }

    Ok(())
}

fn rpassword_prompt(prompt: &str) -> String {
    print!("{}", prompt);
    use std::io::Write;
    std::io::stdout().flush().unwrap();

    // Simple password input (no echo suppression on Windows without extra deps)
    let mut password = String::new();
    std::io::stdin().read_line(&mut password).unwrap();
    password.trim().to_string()
}
