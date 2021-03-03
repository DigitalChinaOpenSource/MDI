## odata 关键字实现进度

### 2020.11.30 更新


| 关键字 | 完成情况 | 含义 |
| ------- | ------- | ------ |
| eq      |done     |等于比较，即equal  |
| lt      |done     |小于，litter than |
| gt      |done     |大于,greater than |
| le      |done     |小于等于,litter than or equal|
| ge      |done     |大于等于，greater than or equal|
| and     |done     |逻辑与|
| or      |done     |逻辑或|
|()       |done     |分组|
|contains(x,y)|done|函数，当字段x中包含y字符时|
|substring(c,x),substring(c,x,y)|done|传两参时，获取c字段从index为x开始的所有子字符串。传三参时，获取c字段从index为x开始之后的y长度子字符串|
|replace(c,x,y)|done|当用y替换c字段中所有的x时|
|indexof(c,x)|done|返回c字段中首次出现x字符串的索引|
|length(c)|done|返回c字段的长度|
|startswith(c,x)|done|返回c字段中以x为开头的信息|
|endswith(c,x)|done|返回c字段中以x为结尾的信息|
|tolower(c)|done|返回c字段全传小写的结果|
|toupper(c)|done|返回c字段全转大写的结果|
|trim(c)|done|返回字段c去除两头的空格的结果|
|concat(c,x)|done|返回字段c与x的连接结果|
|year(c)|done|c为datetime类型，返回c字段的年份|
|years|todo||
|month(c)|done|返回c字段的月份|
|day(c)|done|返回c字段的日期|
|days|todo|
|hour(c)|done|返回c字段的小时|
|minute(c)|done|返回c字段的分钟|
|minutes(c)|todo||
|second(c)|done|返回c字段的秒数|
|seconds(c)|todo|
|isOf()|todo||