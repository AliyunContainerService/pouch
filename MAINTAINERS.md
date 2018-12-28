# Maintainers

## Maintainers List

|GitHub ID| Name | Email|Company|
|:---:| :----:| :---:|:--: |
|[allencloud](https://github.com/allencloud)|Allen Sun|shlallen1990@gmail.com|Alibaba Group|
|[skoo87](https://github.com/skoo87)|Marcky Wu|marckywu@gmail.com|Terminus|
|[Letty5411](https://github.com/Letty5411)|Letty Liu|letty.ll@alibaba-inc.com|Alibaba Group|
|[rudyfly](https://github.com/rudyfly)|Rudy Zhang|rudyflyzhang@gmail.com|Alibaba Group|
|[Ace-Tang](https://github.com/Ace-Tang)|Ace Tang|aceapril@126.com|Alibaba Group|
|[YaoZengzeng](https://github.com/YaoZengzeng)|Zengzeng Yao|yaozengzeng@zju.edu.cn|Zhejiang University|
|[sunyuan3](https://github.com/sunyuan3)|Yuan Sun|yile.sy@alibaba-inc.com |Alibaba Group|
|[HusterWan](https://github.com/HusterWan)|Michael Wan|zirenwan@gmail.com|Alibaba Group|
|[ZYecho](https://github.com/ZYecho)|Yue Zhang|zy675793960@yeah.net|ISCAS|
|[fuweid](https://github.com/fuweid)|Wei Fu|yuge.fw@alibaba-inc.com|Alibaba Group|
|[zhuangqh](https://github.com/zhuangqh)|Jerry Zhuang|zhijin.zqh@alibaba-inc.com| Alibaba Group|

## Component Owner/Backup

In order to make all components in PouchContainer well-maintained, an owner/backup responsibility policy is set for all maintainers. No matter the owner or the backup, they have responsibilities to make/review design and programme code for the corresponding component. They work for success of the decoupled component, and efforts all maintainers do guarantee PouchContainer's success.

We encourage all community participants to communicate with component owner/backup to get more information or guidance via all kinds of channels, such as @him in GitHub issue or pull request, email him via listed email address and so on. And component owner/backups are obligated to provide kind and patient help to the community.

The component owner and backup are listed as below:

|Component|Owner|Backup|Notes|
|:---:|:----:|:---:|:--:|
|CRI/Kubernetes|zhuangqh|Starnop|CRI stablibity/improvement, collaborate with upstream|
|API/CLI|allencloud|fuweid|API/CLI definitions/change reviews|
|runtime/OCI|Ace-Tang| zhuangqh|multiple runtime support, runc/kata/gvisor|
|storage|rudyfly|fuweid|volumes, diskquota, volume plugin protocols|
|network|rudyfly|HusterWan|libnetwork, cni plugins|
|image|fuweid|ZYecho|image storage/distribution/management |
|ctrd|fuweid|Ace-Tang|component used in pouchd to communicated with containerd|
|ContainerMgr|HusterWan|Ace-Tang|container lifecycle management in pouchd|
|Test|sunyuan3|chuanchang|test framework and software quality|
|Document|allencloud|rufyfly|Document Roadmap and quality improvement|
