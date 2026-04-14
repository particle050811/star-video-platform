### hz 脚手架命令

1. **初始化项目**（首次使用）
   ```bash
   hz new --module video-platform --idl idl/user.proto --proto_path=.
   ```

2. **更新/新增模块**（添加新 proto 文件时）
   ```bash
   hz update --idl idl/user.proto --proto_path=.
   hz update --idl idl/video.proto --proto_path=.
   hz update --idl idl/interaction.proto --proto_path=.
   hz update --idl idl/relation.proto --proto_path=.