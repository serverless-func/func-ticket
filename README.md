func-ticket
----
> 12306 购票邮件解析

## 查找邮件中 N 天的车票信息

```shell
curl -X POST "https://ticket.func.dongfg.com/search" -d '{"username": "xxxxx", "password": "yyyyyy", "days": 30}'
```

```json5
{
  "data": [
    {
      "OrderNo": "EB1XXX",
      // 订单号
      "Name": "xxx",
      // 姓名
      "DepartDate": "2025年11月21日",
      // 出发日期
      "DepartTime": "20:54",
      // 出发时间
      "FromStation": "xxx",
      // 出发站
      "ToStation": "xx",
      // 到达站
      "TrainNo": "xx",
      // 车次
      "Seat": "17车2号下铺",
      // 坐席号
      "SeatLevel": "硬卧",
      // 坐席等级
      "Price": 279.5,
      // 票价
      "IsRefund": true,
      // 是否退票
      "RefundFee": 0,
      // 退票费
      "RefundPrice": 279.5
      // 退票金额
    }
  ],
  "msg": "success",
  "timestamp": 1600757374
}
```

### HTTP Request

`POST https://ticket.func.dongfg.com/search`

### JSON Body Fields

| Field    | Default | Description                        |
| -------- | ------- | ---------------------------------- |
| username |         | 邮箱用户名                         |
| password |         | 邮箱密码                           |
| days     | 30      | 解析的邮件范围，可为空，默认 30 天 |