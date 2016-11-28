# 说明

这一份文档应该是放在规则仓库中的，所有的路径都是以规则仓库为根目录，也就是项目中的 `satori-rules`。
如果你是在规则仓库中看到的这个文档，这个说明可以忽略。

# 报警策略设定

报警策略在规则仓库的 `alarm` 目录中配置。
alarm 中的所有 yaml 文件都会解析合并，所以可以任意的拆分文件。

里面的配置都有详细的说明。如果没说明白麻烦提个 issue，或者直接看源码 :)

目前 alarm 支持这几种报警方式：

- [Nexmo](https://www.nexmo.com) 发送电话报警
- [云片](https://www.yunpian.com) 发送短信报警
- [PagerDuty](https://www.pagerduty.com)
- [OneAlert](http://www.onealert.com)
- [微信企业号](https://qy.weixin.qq.com)
- [BearyChat](https://bearychat.com) 里发消息
- SMTP 发邮件
- [BitBar插件](https://getbitbar.com)

如果你实现了其他的方式，欢迎发 PR 回来~

LeanCloud 内部的报警级别是这么设定的：

| LV | 效果                                    |
|----|-----------------------------------------|
| 0  | 语音通话                                |
| 1  | 短信（包括运维手机)                     |
| 2  | 短信                                    |
| 3  | 微信企业号                              |
| 4  | 邮件（但是0123级别的不会发）            |
| 5  | BearyChat                               |
| 6  | 什么都不做（但是会显示在 BitBar 插件中) |

规则仓库的默认配置就是按照这个来的，可以自己按照需求修改。
