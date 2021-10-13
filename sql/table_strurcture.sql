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
CREATE TABLE tResources(
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time',
  ownerPRI varchar(255) comment 'owner',
  alias VARCHAR(255) comment 'alias',
  mimeType VARCHAR(255) comment 'mime type',
  rID VARCHAR(255) NOT NULL comment 'resource ID'
) default charset utf8 comment '';
CREATE TABLE tAccounts(
  id BIGINT UNSIGNED NOT NULL primary key AUTO_INCREMENT comment 'primary key',
  create_time DATETIME COMMENT 'create time',
  update_time DATETIME COMMENT 'update time',
  nickname VARCHAR(255),
  email VARCHAR(255),
  passwd VARCHAR(255)
) default charset utf8 comment '';