/* Copyright (c) 2021 Digital China Group Co.,Ltd
 * Licensed under the GNU General Public License, Version 3.0 (the "License").
 * You may not use this file except in compliance with the license.
 * You may obtain a copy of the license at
 *     https://www.gnu.org/licenses/
 *
 * This program is free; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.0 of the License.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
**/

CREATE TABLE public.user_profile (
  login_name varchar(255) NOT NULL,
  display_name varchar(20) NOT NULL,
  avatar bytea,
  PRIMARY KEY (login_name)
)
;

COMMENT ON COLUMN public.user_profile.login_name IS '邮箱，用作登录名';

COMMENT ON COLUMN public.user_profile.display_name IS '显示名';

COMMENT ON COLUMN public.user_profile.avatar IS '自定义头像的图片';

CREATE TABLE public.project (
  project_id uuid NOT NULL,
  name varchar(50) NOT NULL,
  description varchar(2000) NOT NULL,
  creator varchar(255) NOT NULL,
  create_on timestamp NOT NULL,
  icon bytea,
  PRIMARY KEY (project_id)
)
;

COMMENT ON COLUMN public.project.project_id IS '项目实体的主键';

COMMENT ON COLUMN public.project.name IS '项目名';

COMMENT ON COLUMN public.project.description IS '项目描述';

COMMENT ON COLUMN public.project.creator IS '项目创建者，邮箱';

COMMENT ON COLUMN public.project.create_on IS '创建时间';

COMMENT ON COLUMN public.project.icon IS '图标';

CREATE TABLE public.project_user (
  project_user_id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
  project_id uuid NOT NULL,
  user_login_name varchar(255) NOT NULL,
  is_project_owner bit NOT NULL,
  PRIMARY KEY (project_user_id)
)
;

COMMENT ON COLUMN public.project_user.project_user_id IS '自增主键';

COMMENT ON COLUMN public.project_user.is_project_owner IS '用户是否该项目的管理者。true代表是，false 代表否';

CREATE TABLE public.environment_access_history (
  access_history_id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
  environment_id uuid NOT NULL,
  user_login_name varchar(255) NOT NULL,
  create_on timestamp NOT NULL,
  api_method varchar(20) NOT NULL,
  api_url varchar(2048) NOT NULL,
  http_body varchar(4096),
  PRIMARY KEY (access_history_id)
)
;

COMMENT ON COLUMN public.environment_access_history.access_history_id IS '自增主键';

COMMENT ON COLUMN public.environment_access_history.user_login_name IS '访问授权码的创建者';

COMMENT ON COLUMN public.environment_access_history.create_on IS '创建时间';

COMMENT ON COLUMN public.environment_access_history.api_method IS 'HttpMethod的值';

COMMENT ON COLUMN public.environment_access_history.api_url IS 'Http请求地址';

COMMENT ON COLUMN public.environment_access_history.http_body IS 'Http请求体，可空。如过长则截断';

CREATE TABLE public.environment_history (
  environment_history_id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
  environment_id uuid NOT NULL,
  publisher varchar NOT NULL,
  publish_on timestamp NOT NULL,
  metadata_source xml NOT NULL,
  metadata_target xml NOT NULL,
  action_result xml NOT NULL,
  actions int4 NOT NULL,
  failed_actions int4 NOT NULL,
  PRIMARY KEY (environment_history_id)
)
;

COMMENT ON COLUMN public.environment_history.environment_history_id IS '发布历史Id';

COMMENT ON COLUMN public.environment_history.environment_id IS '所属的Env';

COMMENT ON COLUMN public.environment_history.publisher IS '发布者（邮箱）';

COMMENT ON COLUMN public.environment_history.publish_on IS '发布时间';

COMMENT ON COLUMN public.environment_history.metadata_source IS '发布内容';

COMMENT ON COLUMN public.environment_history.metadata_target IS '发布的目标';

COMMENT ON COLUMN public.environment_history.action_result IS '发布的动作及结果';

COMMENT ON COLUMN public.environment_history.actions IS '冗余字段，表示操作的数量';

COMMENT ON COLUMN public.environment_history.failed_actions IS '冗余字段，表示操作失败的数量';

CREATE TABLE public.environment (
  environment_id uuid NOT NULL,
  project_id uuid NOT NULL,
  type int2 NOT NULL,
  owner varchar(255),
  graph_current xml NOT NULL,
  metadata_current xml NOT NULL,
  metadata_published xml NOT NULL,
  sql_host varchar(255) NOT NULL,
  sql_port int4 NOT NULL,
  sql_user varchar(255) NOT NULL,
  sql_password varchar(255) NOT NULL,
  sql_dbname varchar(255) NOT NULL,
  sql_scheme varchar(255) NOT NULL,
  PRIMARY KEY (environment_id)
)
;

COMMENT ON COLUMN public.environment.environment_id IS '环境的主键';

COMMENT ON COLUMN public.environment.project_id IS '所属的Project';

COMMENT ON COLUMN public.environment.type IS '0表示开发环境，1表示测试环境，2表示生产环境';

COMMENT ON COLUMN public.environment.owner IS '该环境的所有者，仅对开发环境有效。如果是生产或QA环境，则为空。';

COMMENT ON COLUMN public.environment.graph_current IS '当前画布的元数据';

COMMENT ON COLUMN public.environment.metadata_current IS '当前的元数据';

COMMENT ON COLUMN public.environment.metadata_published IS '已发布的元数据';

COMMENT ON COLUMN public.environment.sql_host IS '数据库主机';

COMMENT ON COLUMN public.environment.sql_port IS '数据库端口';

COMMENT ON COLUMN public.environment.sql_user IS '数据库用户';

COMMENT ON COLUMN public.environment.sql_password IS '数据库密码';

COMMENT ON COLUMN public.environment.sql_dbname IS '数据库名';

COMMENT ON COLUMN public.environment.sql_scheme IS '数据库架构名';

CREATE TABLE public.environment_user (
  environment_user_id int8 NOT NULL,
  environment_id uuid NOT NULL,
  user_login_name varchar(255) NOT NULL,
  permission int4 NOT NULL,
  PRIMARY KEY (environment_user_id)
)
;

COMMENT ON COLUMN public.environment_user.environment_user_id IS '自增主键';

COMMENT ON COLUMN public.environment_user.permission IS '位运算：第一位代表是否数据可写，第二位代表可管理和发布模型；即0只读，1可以写入数据但不能修改模型，2代表允许修改模型但是不能写入数据，3代表既可以写数据又可以管理模型';

CREATE TABLE public.environment_token (
  token_id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
  name varchar(255) NOT NULL,
  environment_id uuid NOT NULL,
  user_login_name varchar(255) NOT NULL,
  token varchar(500) NOT NULL,
  create_on timestamp NOT NULL,
  expired_on timestamp,
  remark varchar(255),
  PRIMARY KEY (token_id)
)
;

COMMENT ON COLUMN public.environment_token.token_id IS '自增主键';

COMMENT ON COLUMN public.environment_token.name IS '显示的标题文本';

COMMENT ON COLUMN public.environment_token.user_login_name IS '访问授权码的创建者';

COMMENT ON COLUMN public.environment_token.token IS 'API访问授权码';

COMMENT ON COLUMN public.environment_token.create_on IS '创建时间';

COMMENT ON COLUMN public.environment_token.expired_on IS '失效时间，允许为空，表示长期有效';

COMMENT ON COLUMN public.environment_token.remark IS '备注';

ALTER TABLE public.project_user ADD CONSTRAINT "fk_ProjectUser_Project_1" FOREIGN KEY (project_id) REFERENCES public.project (project_id);

ALTER TABLE public.environment_history ADD CONSTRAINT "fk_EnvironmentHistory_Environment_1" FOREIGN KEY (environment_id) REFERENCES public.environment (environment_id);

ALTER TABLE public.environment ADD CONSTRAINT "fk_Environment_Project_1" FOREIGN KEY (project_id) REFERENCES public.project (project_id);

ALTER TABLE public.environment_user ADD CONSTRAINT "fk_EnvironmentUser_Environment_1" FOREIGN KEY (environment_id) REFERENCES public.environment (environment_id);

ALTER TABLE public.environment_token ADD CONSTRAINT "fk_EnvironmentToken_Environment_1" FOREIGN KEY (environment_id) REFERENCES public.environment (environment_id);