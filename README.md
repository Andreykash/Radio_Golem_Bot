# Radio_Golem_Bot
Шаг 2: Сборка проекта
Создайте новый каталог для вашего проекта:

sh
mkdir -p ~/myproject
cd ~/myproject
Скопируйте код проекта в этот каталог и инициализируйте новый модуль Go:

sh
go mod init myproject
Затем создайте файл main.go и вставьте туда ваш код.

Добавьте все необходимые зависимости в файл go.mod:

sh
go mod tidy
Соберите проект:

sh
go build -o myproject main.go
Шаг 3: Настройка сервиса Systemd
Создайте новый файл сервиса в /etc/systemd/system/:

sh
sudo nano /etc/systemd/system/myproject.service
Вставьте в него следующее содержание:

ini
[Unit]
Description=My Telegram Bot Service
After=network.target

[Service]
ExecStart=/home/yourusername/myproject/myproject
WorkingDirectory=/home/yourusername/myproject
Restart=always
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
Environment=GO_ENV=production

[Install]
WantedBy=multi-user.target
Замените yourusername на ваше имя пользователя и путь на путь к вашему проекту.

Шаг 4: Запуск и управление сервисом
Перезагрузите Systemd для применения изменений:

sh
sudo systemctl daemon-reload
Запустите ваш сервис и включите его автозапуск при загрузке системы:

sh
sudo systemctl start myproject.service
sudo systemctl enable myproject.service
Теперь ваш проект будет запускаться автоматически при старте системы, и вы сможете управлять им с помощью команд Systemd, таких как:

sh
sudo systemctl status myproject.service
sudo systemctl stop myproject.service
sudo systemctl restart myproject.service
Попробуйте эти шаги, и ваш проект должен успешно запуститься как сервис в Ubuntu. Если у вас возникнут вопросы или потребуется дополнительная помощь, дайте знать!
