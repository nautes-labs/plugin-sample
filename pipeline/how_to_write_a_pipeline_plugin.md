## 概述
用户在执行流水线的时可能会希望有前置步骤和后置步骤。为了解决定制化的需求，流水线的前后置步骤通过插件的形式完成。

### 编写插件
nautes 使用的插件系统是 "https://github.com/hashicorp/go-plugin"。可以查看官方文档编写插件。

插件需要实现三个接口(具体参数请查看[protobuf文件](http://github.com/nautes-labs/nautes/app/runtime-operator/pkg/pipeline/proto/hook.proto))
|接口名|接口说明|
|--|--|
|GetPipelineType    |   返回插件支持的pipeline类型。
|GetHooksMetadata   |   返回插件支持的功能的元数据信息（如功能名字，是否可以前置，可选输入参数等）
|BuildHook          |   返回插入到用户流水线前后的代码块。

示例项目存放在 pipeline/sample 目录下，它可以在用户流水线前后 ls 镜像中的指定路径。使用以下命令即可生成插件。
```shell
go env -w CGO_ENABLED=0
go build -o /tmp/sample pipeline/sample/main.go
```

### 添加插件
1. 准备好以下内容:
  - 要添加到 runtime 中的插件。（二进制文件，可执行权限）
  - 一个可以使用 kubectl 访问租户集群的linux环境。
2. 使用以下命令安装插件
    ```bash
    PODNAME=`kubectl -n nautes get po -l control-plane=runtime-operator -o jsonpath='{.items[0].metadata.name}'`
    kubectl cp /tmp/sample nautes/${PODNAME}:/opt/nautes/plugins/sample
    ```
3. 添加完后，可以通过`kubectl -n nautes logs -l control-plane=runtime-operator`在日志中查看添加结果。如果出现以下信息即添加成功
    > plugin: starting plugin: path=/opt/nautes/plugins/sample args=["/opt/nautes/plugins/sample"]

### 删除插件
1. 使用`kubectl -n nautes exec ${PODNAME}  -- rm /opt/nautes/plugins/sample`
2. 删除完成后可以通过查看容器日志查看结果。如果出现以下内容，即删除成功。
    > plugin: plugin process exited: path=plugins/sample

### 使用插件
1. 在 project pipeline runtime 文件中添加以下片段。
```yaml
spec:
  hooks:
    preHooks:
      - name: ls            # 插件中定义的hook的名字。
        vars:               # 插件中定义的可输入的变量。
          imageName: bash
          printPath: /var
```