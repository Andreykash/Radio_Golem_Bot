# Radio_Golem_Bot
**Сборка проекта**
**Шаг 1: Создайте новый каталог для вашего проекта:**

    sh
    mkdir -p ~/radio_golem
    cd ~/radio_golem

Скопируйте код проекта в этот каталог и инициализируйте новый модуль Go:

    sh
    go mod init radio_golem

**Шаг 2: Затем создайте файл main.go и вставьте туда ваш код.**

Добавьте все необходимые зависимости в файл go.mod:

    sh
    go mod tidy

Соберите проект:

    sh
    go build -o radio_golem main.go

**Шаг 3: Настройка сервиса Systemd**
Создайте новый файл сервиса в /etc/systemd/system/:

    sh
    sudo nano /etc/systemd/system/radio_golem.service

Вставьте в него следующее содержание:

    ini
    [Unit]
    Description=My radio_golem Bot Service
    After=network.target
    
    [Service]
    ExecStart=/home/yourusername/myproject/radio_golem
    WorkingDirectory=/home/yourusername/radio_golem
    Restart=always
    Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
    Environment=GO_ENV=production
    
    [Install]
    WantedBy=multi-user.target
    Замените yourusername на ваше имя пользователя и путь на путь к вашему проекту.

**Шаг 4: Запуск и управление сервисом**
Перезагрузите Systemd для применения изменений:

    sh
    sudo systemctl daemon-reload

Запустите ваш сервис и включите его автозапуск при загрузке системы:

    sh
    sudo systemctl start radio_golem.service
    sudo systemctl enable radio_golem.service

Теперь ваш проект будет запускаться автоматически при старте системы, и вы сможете управлять им с помощью команд Systemd, таких как:

    sh
    sudo systemctl status radio_golem.service
    sudo systemctl stop radio_golem.service
    sudo systemctl restart radio_golem.service

Попробуйте эти шаги, и ваш проект должен успешно запуститься как сервис в Ubuntu.
