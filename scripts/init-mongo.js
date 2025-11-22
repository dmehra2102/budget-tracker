// MongoDB initialization script
db = db.getSiblingDB("budget_tracker");

// Create collections
db.createCollection("users");
db.createCollection("budgets");
db.createCollection("expenses");
db.createCollection("alerts");
db.createCollection("alert_notifications");

// Create indexes
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ reset_token: 1 });

db.budgets.createIndex({ user_id: 1, start_date: -1 });
db.budgets.createIndex({ is_active: 1, end_date: 1 });

db.expenses.createIndex({ budget_id: 1, date: -1 });
db.expenses.createIndex({ user_id: 1, date: -1 });

db.alerts.createIndex({ user_id: 1 });
db.alerts.createIndex({ budget_id: 1 });
db.alerts.createIndex({ is_enabled: 1 });

db.alert_notifications.createIndex({ user_id: 1, sent_at: -1 });

print("Database initialized successfully");
