
骆又鸣
💬百技培训ing
骆又鸣(阿里健康科技（中国）有限公司-引擎&算法-医疗大数据)
今天 09:20
过来扫码
速来速来 不然整租表演lol
教室怎么走最快
已读
都挺快的 就在六的后面
我说进了7号后
已读
二楼就能看到
就那个开着门的教师
电梯口
等电梯中
已读
今天 10:12
你配好了么
mei
已读
是不是要垮
不慌
已读
今天 17:24
https://github.com/alibaba/pouch/blob/master/docs/contributions/code_styles.md
已读
https://jbt.github.io/markdown-editor/
37分钟前
# Code Style 编码规范
 
Code style is a set of rules or guidelines when writing source codes of a software project. Following particular code style will definitely help contributors to read and understand source codes very well. In addition, it will help to avoid introducing errors as well.

编码规范是一组写软件项目源代码的规则和指南。遵守一定的规范会帮助开源贡献者们更好地阅读和理解源代码。此外，它也有助于避免产生错误。

## Code Style Tools

Project PouchContainer is written in Golang. And currently we use three tools to help conform code styles in this project. These three tools are:

* [gofmt](https://golang.org/cmd/gofmt)
* [golint](https://github.com/golang/lint)
* [go vet](https://golang.org/cmd/vet/)

And all these tools are used in [Makefile](../../Makefile).

## Code Review Comments 代码审查注释

When collaborating in PouchContainer project, we follow the style from [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments). Before contributing, we treat this as a must-read.

在PouchContainer项目中进行协作时，我们遵循[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)的规则。在开发之前，我们将这些建议视为必读。

## Additional Style Rules 额外的规范

For a project, existing tools and rules may not be sufficient. To align more in styles, we recommend contributors taking a thorough look at the following additional style rules:

对于一个项目来说，现有的工具和规则可能并不够。为了更好的整合样式，我们推荐贡献者仔细阅读以下一些额外的规则：

### RULE001 - Add blank line between field's comments 规则001 - 在字段（fields）注释之间添加空行

When constructing a struct, if comments needed for fields in struct, keep a blank line between fields. The encouraged way is as following:
当构建一个结构（struct）时，如果结构中的字段需要注释，在字段之间增加一行空行。我们推荐如下的方式：

``` golang
// correct example
// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
  // Store stores containers in Backend store.
  // Element operated in store must has a type of *ContainerMeta.
  // By default, Store will use local filesystem with json format to store containers.
  Store *meta.Store

  // Client is used to interact with containerd.
  Client ctrd.APIClient

  // NameToID stores relations between container's name and ID.
  // It is used to get container ID via container name.
  NameToID *collect.SafeMap
  ......
}
```
我们不推荐以下的方式：
Rather than:

```golang
// wrong example
// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
  // Store stores containers in Backend store.
  // Element operated in store must has a type of *ContainerMeta.
  // By default, Store will use local filesystem with json format to store containers.
  Store *meta.Store
  // Client is used to interact with containerd.
  Client ctrd.APIClient
  // NameToID stores relations between container's name and ID.
  // It is used to get container ID via container name.
  NameToID *collect.SafeMap
  ......
}
```

### RULE002 - Add parameter name in interface definition 规则002 - 在接口（interface）定义中增加参数名（parameter）

When defining interface functions, we should always explicitly add formal parameters, and this helps a lot to code readability. For example, the following way are preferred:
当定义接口时，我们应该明确地添加形式参数（formal parameters），这将会显著的提高代码可读性。比如，我们推荐下面的写法：



``` golang
// correct example
// ContainerMgr is an interface to define all operations against container.
type ContainerMgr interface {
  // Start a container.
  Start(ctx context.Context, id, detachKeys string) error

  // Stop a container.
  Stop(ctx context.Context, name string, timeout int64) error
  ......
}
```

However, missing formal parameter's name would make interface unreadable, since we would never know what the parameter's real meaning unless turning to one implementation of this interface:
但是，缺失的形式参数名会让接口不可读，因为我们不会知道参数的真正含义，除非我们实现这个接口。

``` golang
// wrong example
type ContainerMgr interface {
  // Start a container.
  Start(context.Context, string, string) error

  // Stop a container.
  Stop(context.Context, string, int64) error
  ......
}

```

In addition, a blank line between function's comments is encouraged to make interface more readable.

### RULE003 - Import Packages 规则003 - 倒入包

When importing packages, to improve readabilities, we should import package by sequence:
当倒入包时，为了提高可读性，我们应该按如下顺序倒入：

* Golang's built-in system packages; Golang的内置系统包
* project's own packages; 项目自己的包
* third-party packages. 第三方的包

And we should keep a blank line among these three kinds of packages like the following:
我们应该在这三类包之间加入一个空行，如下所示：

``` golang
import (
  "fmt"
  "strconv"
  "strings"

  "github.com/alibaba/pouch/apis/types"
  "github.com/alibaba/pouch/pkg/errtypes"
  "github.com/alibaba/pouch/pkg/meta"
  "github.com/alibaba/pouch/pkg/randomid"

  "github.com/opencontainers/selinux/go-selinux/label"
  "github.com/pkg/errors"
)
```

### RULE004 - Variable declaration position 规则004 - 变量声明位置

Variable object should be declared at the beginning of the go file following package name and importing. 变量对象应该放在go文件的开头部位，在包名称和倒入之后的位置。

### RULE005 - Generation of action failure 规则005 - 生成错误信息

When generating error in one function execution failure, we should generally use the following way to append string "failed to do something" and the specific err instance to construct a new error:当一个函数运行失败产生错误时，通常来说，我们应该用如下的方式将字段 "failed to do something" 和具体的错误结合在一起去构造错误信息。
``` golang
fmt.Errorf("failed to do something: %v", err)
```

When an err could be thrown out, please remember to add it in the error construction.
当一个错误被抛出时，应该记得将它加到错误构造中。
long_text_2018-07-09-19-58-06.txt
6.7 KBTXT
预览
打开
点赞
评论
已查收
https://github.com/alibaba/pouch/blob/master/ROADMAP.md
已读
编码规范化工具
PouchContainer项目是用Go语言编写的。目前我们使用三个工具来帮助整合此项目中的代码样式。这三个工具是：
所有这些工具都有在[Makefile](../../Makefile)中使用。
已读
刚刚
20:34
# 编码规范

编码规范是一系列编写开源项目源代码的规则和指南。遵守一定的规范会帮助开源贡献者们更好地阅读和理解源代码。此外，它也有助于避免产生错误。

## 编码规范化工具

PouchContainer项目是用Go语言编写的。目前我们使用三个工具来帮助整合此项目中的代码样式。这三个工具是：

* [gofmt](https://golang.org/cmd/gofmt)
* [golint](https://github.com/golang/lint)
* [go vet](https://golang.org/cmd/vet/)

所有这些工具都有在[Makefile](../../Makefile)中使用。

## 代码审查注释

在PouchContainer项目中进行协作时，我们遵循[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)的规范。在开发之前，我们将这些建议视为必读。

## 额外的规范

对于一个项目来说，现有的工具和规则可能并不够。为了更好的整合样式，我们推荐贡献者仔细阅读以下一些额外的规范：

### 规则001 - 在字段（fields）注释之间添加空行

当构建一个结构（struct）时，如果结构中的字段需要注释，需在字段之间增加一行空行。我们推荐如下的方式：

``` golang
// correct example
// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
  // Store stores containers in Backend store.
  // Element operated in store must has a type of *ContainerMeta.
  // By default, Store will use local filesystem with json format to store containers.
  Store *meta.Store

  // Client is used to interact with containerd.
  Client ctrd.APIClient

  // NameToID stores relations between container's name and ID.
  // It is used to get container ID via container name.
  NameToID *collect.SafeMap
  ......
}
```

我们不推荐以下的方式：

```golang
// wrong example
// ContainerManager is the default implement of interface ContainerMgr.
type ContainerManager struct {
  // Store stores containers in Backend store.
  // Element operated in store must has a type of *ContainerMeta.
  // By default, Store will use local filesystem with json format to store containers.
  Store *meta.Store
  // Client is used to interact with containerd.
  Client ctrd.APIClient
  // NameToID stores relations between container's name and ID.
  // It is used to get container ID via container name.
  NameToID *collect.SafeMap
  ......
}
```

### 规则002 - 在接口（interface）定义中增加参数名（parameter）

当定义接口时，我们应该明确地添加形式参数（formal parameters），这将会显著的提高代码可读性。比如，我们推荐下面的写法：

``` golang
// correct example
// ContainerMgr is an interface to define all operations against container.
type ContainerMgr interface {
  // Start a container.
  Start(ctx context.Context, id, detachKeys string) error

  // Stop a container.
  Stop(ctx context.Context, name string, timeout int64) error
  ......
}
```

但是，缺失的形式参数名会让接口不可读，因为我们不会知道参数的真正含义，除非我们实现这个接口。以下是一个错误事例：


``` golang
// wrong example
type ContainerMgr interface {
  // Start a container.
  Start(context.Context, string, string) error

  // Stop a container.
  Stop(context.Context, string, int64) error
  ......
}

```
除此之外，在不同方法（function）的注释之间应该留一个空行以增强可读性。

### 规则003 - 倒入包

当倒入包时，为了提高可读性，我们应该按如下顺序倒入：

* Golang的内置系统包;
* 项目自己的包；
* 第三方的包；

我们应该在这三类包之间加入一个空行，如下所示：

``` golang
import (
  "fmt"
  "strconv"
  "strings"

  "github.com/alibaba/pouch/apis/types"
  "github.com/alibaba/pouch/pkg/errtypes"
  "github.com/alibaba/pouch/pkg/meta"
  "github.com/alibaba/pouch/pkg/randomid"

  "github.com/opencontainers/selinux/go-selinux/label"
  "github.com/pkg/errors"
)
```

### 规则004 - 变量声明位置

变量对象应该放在go文件的开头部位，在包名称和倒入之后的位置。

### 规则005 - 生成错误信息

当一个函数运行失败产生错误时，通常来说，我们应该用如下的方式将字段 "failed to do something" 和具体的错误结合在一起去构造错误信息，比如：

``` golang
fmt.Errorf("failed to do something: %v", err)
```

当一个错误被抛出时，应该将它加到错误构造中。

### 规则006 - 及早返回以减少缩进

PouchContainer鼓励贡献者通过“及早返回”简单化源代码，从而减少缩进。例如，这样的编写方式是不被鼓励的：

``` golang
// wrong example
if retry {
  if t, err := calculateSleepTime(d); err == nil {
    time.Sleep(t)
    times++
    return retryLoad()
  }
  return fmt.Errorf("failed to calculate timeout: %v", err)
}
return nil
```


以上代码中有一些可以避免的缩进。我们鼓励以下的编写方式：

``` golang
// correct example
if !retry {
  return nil
}

t, err := calculateSleepTime(d);
if err != nil {
  return fmt.Errorf("failed to calculate timeout: %v", err)
}

time.Sleep(t)
times++

return retryLoad()
```

### 规则007 - 日志与错误使用小写字母写

无论是日志还是错误，消息的第一个字母应小写。即，我们鼓励`logrus.Debugf("failed to add list: %v", err)`，而不鼓励`logrus.Debugf("Failed to add list: %v", err)`。

### 规则008 - 嵌套错误

嵌套错误建议使用`github.com/pkg/errors`包处理。


### 规则009 - 用正确的标注格式

无论是关于变量、结构、函数、代码块还是其他任何元素的注释都以`//`加一个空格开头。切记不要忘记空格，并用`.`结束所有句子。此外，函数的注释中我们鼓励使用第三人称单数。例如，以下方式：

```golang
// wrong example
// ExecContainer execute a process in container.
func (c *Client) ExecContainer(ctx context.Context, process *Process) error {
  ......
}
```

中有语法错误。应把`executes`改成`execute`：
could -> has to
could be polished to be `executes` rather than `execute`:

```golang
// correct example
// ExecContainer executes a process in container.
func (c *Client) ExecContainer(ctx context.Context, process *Process) error {
  ......
}
```

### 规则010 - 切忌重复工作

我们在编程中应时刻铭记`DRY(Don't Repeat Yourself)`。

### 规则011 - 我们重视您的意见

如果您觉得我们的编码规范有不足之处，我们将非常感谢您提交pull request来帮助我变好。
