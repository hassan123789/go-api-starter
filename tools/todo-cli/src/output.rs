use anyhow::Result;
use colored::Colorize;

use crate::api::Todo;

pub fn print_todos(todos: &[Todo], format: &str) -> Result<()> {
    match format {
        "json" => {
            println!("{}", serde_json::to_string_pretty(todos)?);
        }
        _ => {
            if todos.is_empty() {
                println!("{}", "No todos found.".dimmed());
                return Ok(());
            }

            println!("{}", format!("📋 {} todos:", todos.len()).bold());
            println!();

            for todo in todos {
                print_todo_line(todo);
            }
        }
    }
    Ok(())
}

pub fn print_todo(todo: &Todo, format: &str) -> Result<()> {
    match format {
        "json" => {
            println!("{}", serde_json::to_string_pretty(todo)?);
        }
        _ => {
            print_todo_detail(todo);
        }
    }
    Ok(())
}

fn print_todo_line(todo: &Todo) {
    let status = if todo.completed {
        "✓".green()
    } else {
        "○".yellow()
    };

    let title = if todo.completed {
        todo.title.strikethrough().dimmed().to_string()
    } else {
        todo.title.clone()
    };

    println!("  {} #{} {}", status, todo.id.to_string().dimmed(), title);
}

fn print_todo_detail(todo: &Todo) {
    let status = if todo.completed {
        "Completed".green()
    } else {
        "Pending".yellow()
    };

    println!("{}", "─".repeat(40).dimmed());
    println!("  {} #{}", "Todo".bold(), todo.id);
    println!("  {}: {}", "Title".dimmed(), todo.title);
    println!("  {}: {}", "Status".dimmed(), status);
    println!("  {}: {}", "Created".dimmed(), format_datetime(&todo.created_at));
    println!("  {}: {}", "Updated".dimmed(), format_datetime(&todo.updated_at));
    println!("{}", "─".repeat(40).dimmed());
}

fn format_datetime(dt: &str) -> String {
    // Try to parse and format nicely, fallback to original
    chrono::DateTime::parse_from_rfc3339(dt)
        .map(|d| d.format("%Y-%m-%d %H:%M").to_string())
        .unwrap_or_else(|_| dt.to_string())
}
