; GoStudy Windows 安装包（Inno Setup 6.5+）
;
; 编译命令（本地或 CI 都走 app/scripts/build-windows-installer.sh）：
;   ISCC.exe GoStudy.iss /DAppVersion=0.5.0 /DSourceExe=..\..\release\GoStudy.exe
;
; 产物：app/release/GoStudy-Setup.exe
; 做出来的安装包会：让用户选安装目录、问要不要桌面快捷方式、在开始菜单建目录、
; 在「应用和功能」里注册卸载项，并在装完后提供「立即打开 GoStudy」。
; 学习进度不在安装目录里（在 %AppData%\go-study），所以卸载不会动用户数据。

#define AppName "GoStudy"
#define AppPublisher "xianshujun"
#define AppURL "https://github.com/xianshujun/awsome-go-study"
#define AppExeName "GoStudy.exe"

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef SourceExe
  #define SourceExe "..\..\release\" + AppExeName
#endif
; VersionInfoVersion 要四段数字，0.5.0 补成 0.5.0.0
#define AppVersionQuad AppVersion + ".0"

[Setup]
; AppId 是升级与卸载的识别标志，一旦发布就不能改：改了会被当成另一个程序重复安装。
AppId={{FB9FE764-6145-4C0B-A69A-EC7730ADDC10}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
AppSupportURL={#AppURL}/issues
AppUpdatesURL={#AppURL}/releases
VersionInfoVersion={#AppVersionQuad}
VersionInfoProductVersion={#AppVersion}
VersionInfoProductName={#AppName}
VersionInfoDescription={#AppName} {#AppVersion} 安装程序
VersionInfoCompany={#AppPublisher}
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
AllowNoIcons=yes
; 默认按「仅为我安装」走，不弹 UAC；在向导里可以改成「为所有用户安装」（那时会装到 Program Files）。
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0
OutputDir=..\..\release
OutputBaseFilename=GoStudy-Setup
SetupIconFile=gostudy.ico
UninstallDisplayName={#AppName}
UninstallDisplayIcon={app}\{#AppExeName}
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
; 装的时候如果 GoStudy 正在运行，先让用户关掉再替换文件
CloseApplications=yes

[Languages]
; 简体中文语言文件是 Inno Setup 6.5 起随编译器分发的；老版本编译器就退回英文界面。
#if FileExists(AddBackslash(CompilerPath) + "Languages\ChineseSimplified.isl")
Name: "chinese"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"
#endif
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: checkedonce

[Files]
Source: "{#SourceExe}"; DestDir: "{app}"; DestName: "{#AppExeName}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\{#AppExeName}"
Name: "{group}\{cm:UninstallProgram,{#AppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExeName}"; Tasks: desktopicon

[Run]
; 以原始用户身份启动：选「为所有用户安装」时安装程序是提权的，直接启动会变成管理员身份，
; 那样应用会把进度写进管理员账户的 %AppData%。
Filename: "{app}\{#AppExeName}"; Description: "{cm:LaunchProgram,{#AppName}}"; Flags: nowait postinstall skipifsilent runasoriginaluser
