; 自定义NSIS脚本

; 在安装前执行
Function .onInit
  ; 可以在这里添加自定义初始化代码
  ; 例如：检查系统要求、创建临时目录等
FunctionEnd

; 在安装后执行
Function .onInstSuccess
  ; 可以在这里添加安装成功后的自定义代码
  ; 例如：注册服务、启动应用等
FunctionEnd

; 在卸载前执行
Function un.onInit
  ; 可以在这里添加卸载前的自定义代码
  ; 例如：停止服务、备份数据等
FunctionEnd

; 在卸载后执行
Function un.onUninstSuccess
  ; 可以在这里添加卸载成功后的自定义代码
  ; 例如：清理注册表、删除残留文件等
FunctionEnd