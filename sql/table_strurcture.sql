-- core internal
CREATE TABLE tPermissionGroups(
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time',
  memberPRI VARCHAR(255) comment 'member',
  gid VARCHAR(255) NOT NULL comment 'gid'
) default charset utf8 comment '';
CREATE TABLE tPermissions(
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time',
  which TINYINT UNSIGNED comment 'which',
  takerPRI VARCHAR(255) comment 'taker',
  resourcePRI VARCHAR(255) comment 'resource',
  kind VARCHAR(255) COMMENT 'kind'
) default charset utf8 comment '';


CREATE TABLE tAccounts(
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time',
  nickname VARCHAR(255),
  email VARCHAR(255),
  passwd VARCHAR(255)
) default charset utf8 comment '';
CREATE TABLE tResources(
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time',
  ownerPRI varchar(255) comment 'owner',
  alias VARCHAR(255) comment 'alias',
  mimeType VARCHAR(255) comment 'mime type',
  quotaUsage VARCHAR(255) comment 'quota usage',
  rID VARCHAR(255) NOT NULL comment 'resource ID'
) default charset utf8 comment '';
CREATE TABLE tQuotas(
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time',
  targetPRI varchar(255) comment 'targetPRI',
  capcity BIGINT UNSIGNED NOT NULL DEFAULT 0 comment 'capcity',
  used BIGINT UNSIGNED NOT NULL DEFAULT 0 comment 'used'
) default charset utf8 comment '';

-- extra

-- Closure Table for Tree. Query: 😄, Move: 😐.
-- Ref from https://kyle.ai/blog/6905.html
CREATE TABLE tHierarchy(
  ancestor VARCHAR(255) NOT NULL comment 'ancestor',
  descendant VARCHAR(255) NOT NULL comment 'descendant',
  distance BIGINT UNSIGNED NOT NULL comment 'distance',
  targetPRI VARCHAR(255) NOT NULL comment 'user PRI',
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time'
) default charset utf8 comment '';
CREATE TABLE tHierarchyData(
  `hierarchyID` VARCHAR(255) NOT NULL comment 'hierarchy ID',
  `name` VARCHAR(255) NOT NULL,
  `resourcePRI` VARCHAR(255),
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time'
) default charset utf8 comment '';