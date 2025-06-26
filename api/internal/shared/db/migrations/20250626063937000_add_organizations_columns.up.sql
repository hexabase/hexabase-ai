-- organizationsテーブルにDisplayName, Description, Website, Email, Status, OwnerID, DeletedAtを追加
ALTER TABLE organizations ADD COLUMN display_name VARCHAR(255);
ALTER TABLE organizations ADD COLUMN description TEXT;
ALTER TABLE organizations ADD COLUMN website VARCHAR(255);
ALTER TABLE organizations ADD COLUMN email VARCHAR(255);
ALTER TABLE organizations ADD COLUMN status VARCHAR(32) DEFAULT 'active';
ALTER TABLE organizations ADD COLUMN owner_id VARCHAR(64);
ALTER TABLE organizations ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE NULL;
-- statusカラムのデフォルト値をactiveに設定
ALTER TABLE organizations ALTER COLUMN status SET DEFAULT 'active';
