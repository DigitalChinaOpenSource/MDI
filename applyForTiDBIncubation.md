# Incubating Projarm
Modelized Data Interface（下文简称，MDI）
Modelized Data Interface(Hereinafter referred to as, MDI)

# Describe the feature or project you want to incubate:
## Summary
我们注意到，当前在TiDB的生态中还没有专门的数据服务工具，即：通过已有的数据库，方便的获取高性能API以完成对数据的操作（或访问）以达到简化（或代替）后端开发工作的目标，不用再重复为第三方系统开发接口这种低效工作。
We note that there is no dedicated data service tool in the TiDB ecosystem, i.e., it is easy to obtain high-performance APIs for data manipulation (or access) through existing databases in order to achieve the goal of simplifying (or replacing) the back-end development work without repeating the inefficient work of developing interfaces for third-party systems.

## Motivation
此项目主要解决以下问题：
- 数据集中管理。TiDB天然适合存储大量数据并成为数据中心，那么这些数据的管理是需要有专门的工具的。MDI目前设计的功能包括数据模型及其版本的管理、模型关系的管理、数据访问权限的管理。
- 灵活的数据查询。当前市场上Automatic API generation Tool大都是生成RestFul这种传统API，灵活性较差、可用性较低。那么MDI还支持GraphQL、OData这样的自定义查询，提高系统查询的灵活性和API的可用性，真正减少后端开发工作量。另外，为了支持更复杂的查询，可以考虑SQL封装功能，将一个段SQL查询封装成API。
- 灵活的数据处理。API的组合串联功能。通过排列增删改查的API来封装一个简单的业务逻辑并成为一个RestFul API，同时具备一定的事务性。
- 触发器的替代方案。TiDB目前还不支持触发器，相关的功能实现需要用户自己在业务层面操作，比较繁琐，MDI可以提供统一的URL回调功能，实现类似触发器的效果。

This project mainly addresses the following issues.
- Centralized data management. TiDB is naturally suitable for storing large amount of data and becoming a data center, then the management of these data is required to have special tools. the functions currently designed by MDI include the management of data models and their versions, the management of model relationships, and the management of data access rights.
- Flexible data query. Most of the Automatic API generation Tool on the market currently generates traditional APIs like RestFul, which is less flexible and less available. Then MDI also supports custom queries like GraphQL and OData to improve the flexibility of system queries and the availability of APIs, and really reduce the back-end development workload. In addition, in order to support more complex queries, you can consider SQL wrapping function to wrap a segment SQL query into an API.
- Flexible data processing. combined crosstalk function of APIs. Encapsulate a simple business logic and become a RestFul API by arranging add, delete, and check APIs with some transactivity.
- TiDB currently does not support triggers, and the implementation of related functions requires users to operate at the business level, which is tedious.

## Estimated Time
6 Months

## RFC/Proposal