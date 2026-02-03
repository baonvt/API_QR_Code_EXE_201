-- Migration: Clean all tables, menu items, and categories for all accounts
-- This will delete all data from tables, menu_items, and categories

-- Delete all order items first (foreign key constraint)
DELETE FROM order_items;

-- Delete all orders
DELETE FROM orders;

-- Delete all tables
DELETE FROM tables;

-- Delete all menu items
DELETE FROM menu_items;

-- Delete all categories
DELETE FROM categories;

-- Reset identity columns (SQL Server)
-- DBCC CHECKIDENT ('tables', RESEED, 0);
-- DBCC CHECKIDENT ('menu_items', RESEED, 0);
-- DBCC CHECKIDENT ('categories', RESEED, 0);
-- DBCC CHECKIDENT ('orders', RESEED, 0);
-- DBCC CHECKIDENT ('order_items', RESEED, 0);
