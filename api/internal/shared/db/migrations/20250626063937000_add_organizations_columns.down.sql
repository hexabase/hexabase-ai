-- organizationsテーブルからDisplayName, Description, Website, Email, Status, OwnerID, DeletedAtを削除
ALTER TABLE organizations DROP COLUMN IF EXISTS display_name;
ALTER TABLE organizations DROP COLUMN IF EXISTS description;
ALTER TABLE organizations DROP COLUMN IF EXISTS website;
ALTER TABLE organizations DROP COLUMN IF EXISTS email;
ALTER TABLE organizations DROP COLUMN IF EXISTS status;
ALTER TABLE organizations DROP COLUMN IF EXISTS owner_id;
ALTER TABLE organizations DROP COLUMN IF EXISTS deleted_at;
