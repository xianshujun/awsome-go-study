# 下载并安装

按照以下步骤快速下载并安装 Go。

关于安装的其他内容，你可能也会感兴趣：
- [管理 Go 安装](https://go.dev/doc/manage-install) —— 如何安装多个版本以及如何卸载。
- [从源码安装 Go](https://go.dev/doc/install/source) —— 如何检出源代码、在自己的机器上构建并运行。

## Go 安装方法

请选择你计算机操作系统的对应选项卡，然后按照其安装说明进行操作。

---

### **macOS**

#### 安装要求
支持 macOS 10.15 及更高版本。

#### 安装步骤
1. 下载 [Go 的 macOS 安装包](https://go.dev/dl/)（`.pkg` 格式）。
2. 双击下载的包文件并按照提示安装 Go。
    - 安装程序会将 Go 发行版安装到 `/usr/local/go` 目录。
    - 安装程序应自动将 `/usr/local/go/bin` 添加到你的 `PATH` 环境变量中。
3. 安装完成后，你可能需要关闭并重新打开任何已打开的终端窗口，以使环境变量的更改生效。

#### 验证安装
1. 打开终端（Terminal）。
2. 输入以下命令：
   ```bash
   $ go version
   ```
3. 确认命令输出显示了你安装的 Go 版本，例如：
   ```
   go version go1.24.5 darwin/amd64
   ```

---

### **Linux**

#### 安装要求
支持 Intel 或 AMD 的 64 位处理器（x86-64）。

#### 安装步骤
1. 移除旧版本的 Go（如果存在）：
   ```bash
   $ rm -rf /usr/local/go
   ```
2. 下载适用于 Linux 的 Go 压缩包（`.tar.gz` 格式）。
3. 将下载的压缩包解压到 `/usr/local`，以创建一个新的 Go 树结构：
   ```bash
   $ tar -C /usr/local -xzf go1.24.5.linux-amd64.tar.gz
   ```
   > **注意**：
   > - 你可能需要使用 `sudo` 权限分别执行上述命令。
   > - **不要**将压缩包解压到已存在的 `/usr/local/go` 目录中，这会导致 Go 安装损坏。

4. 将 `/usr/local/go/bin` 添加到 `PATH` 环境变量中：
    - 可以通过将以下行添加到你的 `$HOME/.profile` 或 `/etc/profile`（系统级安装）中实现：
      ```bash
      export PATH=$PATH:/usr/local/go/bin
      ```
   > **注意**：对 profile 文件的更改可能需要在你下次登录系统后才生效。若要立即生效，可直接运行该命令，或使用如下命令加载配置：
   > ```bash
   > source $HOME/.profile
   > ```

#### 验证安装
1. 打开终端。
2. 输入以下命令：
   ```bash
   $ go version
   ```
3. 确认命令输出显示了你安装的 Go 版本，例如：
   ```
   go version go1.24.5 linux/amd64
   ```

---

### **Windows**

#### 安装要求
支持 Windows 10 及更高版本。

#### 安装步骤
1. 下载 [Go 的 Windows 安装程序](https://go.dev/dl/)（`.msi` 格式）。
2. 双击下载的 `.msi` 文件，并按照提示安装 Go。
    - 默认情况下，安装程序会将 Go 安装到 `Program Files` 或 `Program Files (x86)` 目录下。
    - 你可以根据需要更改安装路径。
3. 安装完成后，你需要关闭并重新打开任何已打开的命令提示符（Command Prompt）或 PowerShell 窗口，以便使安装程序对环境变量的更改生效。

#### 验证安装
1. 在 **Windows** 中，点击 **开始** 菜单。
2. 在搜索框中输入 `cmd`，然后按 **Enter** 键。
3. 在打开的命令提示符窗口中，输入以下命令：
   ```cmd
   go version
   ```
4. 确认命令输出显示了你安装的 Go 版本，例如：
   ```
   go version go1.24.5 windows/amd64
   ```

---

### 验证安装成功
无论使用哪种操作系统，安装完成后都应运行 `go version` 命令来确认 Go 已正确安装，并查看当前安装的版本号。

如需进一步学习，请参考 [Go 入门教程](https://go.dev/doc/tutorial/getting-started)。