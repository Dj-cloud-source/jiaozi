CREATE TABLE users (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    phone           VARCHAR(20) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    nickname        VARCHAR(50) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_users_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE admins (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    username        VARCHAR(50) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_admins_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE stations (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_stations_name (name),
    KEY idx_stations_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE trains (
    id                      BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    train_no                VARCHAR(10) NOT NULL,
    departure_date          DATE NOT NULL,

    departure_station_id    BIGINT UNSIGNED NOT NULL,
    arrival_station_id      BIGINT UNSIGNED NOT NULL,

    departure_time          DATETIME NOT NULL,
    arrival_time            DATETIME NOT NULL,
    sale_start_time         DATETIME NOT NULL,

    first_class_price       DECIMAL(10,2) NULL,
    second_class_price      DECIMAL(10,2) NULL,

    first_class_seat_count  INT UNSIGNED NOT NULL DEFAULT 0,
    second_class_seat_count INT UNSIGNED NOT NULL DEFAULT 0,

    status                  VARCHAR(20) NOT NULL DEFAULT 'DRAFT',

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                            ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_train_date (
        train_no,
        departure_date
    ),

    KEY idx_train_route_date (
        departure_station_id,
        arrival_station_id,
        departure_date
    ),

    KEY idx_train_sale_start_time (sale_start_time),
    KEY idx_train_departure_time (departure_time),
    KEY idx_train_status (status),

    CONSTRAINT fk_train_departure_station
        FOREIGN KEY (departure_station_id)
        REFERENCES stations(id),

    CONSTRAINT fk_train_arrival_station
        FOREIGN KEY (arrival_station_id)
        REFERENCES stations(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE seats (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    train_id        BIGINT UNSIGNED NOT NULL,

    seat_class      VARCHAR(20) NOT NULL,
    seat_no         INT UNSIGNED NOT NULL,

    status          VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE',

    locked_order_id CHAR(36) NULL,
    locked_at       DATETIME NULL,

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_train_class_seat (
        train_id,
        seat_class,
        seat_no
    ),

    KEY idx_seat_available (
        train_id,
        seat_class,
        status,
        seat_no
    ),

    CONSTRAINT fk_seat_train
        FOREIGN KEY (train_id)
        REFERENCES trains(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE passengers (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    user_id         BIGINT UNSIGNED NOT NULL,

    name            VARCHAR(100) NOT NULL,
    id_card         VARCHAR(30) NOT NULL,

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_passenger_user (user_id),
    KEY idx_passenger_id_card (id_card),

    CONSTRAINT fk_passenger_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ticket_orders (
    id                  CHAR(36) PRIMARY KEY,

    user_id             BIGINT UNSIGNED NOT NULL,
    passenger_id        BIGINT UNSIGNED NOT NULL,
    train_id            BIGINT UNSIGNED NOT NULL,
    seat_id             BIGINT UNSIGNED NOT NULL,

    ticket_price        DECIMAL(10,2) NOT NULL,

    status              VARCHAR(30) NOT NULL,

    payment_deadline    DATETIME NOT NULL,

    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ticketed_at         DATETIME NULL,
    cancelled_at        DATETIME NULL,
    returned_at         DATETIME NULL,
    completed_at        DATETIME NULL,

    KEY idx_order_user_created (
        user_id,
        created_at
    ),

    KEY idx_order_train (
        train_id
    ),

    KEY idx_order_status (
        status
    ),

    KEY idx_order_payment_deadline (
        status,
        payment_deadline
    ),

    KEY idx_order_passenger_train (
        passenger_id,
        train_id
    ),

    CONSTRAINT fk_order_user
        FOREIGN KEY (user_id)
        REFERENCES users(id),

    CONSTRAINT fk_order_passenger
        FOREIGN KEY (passenger_id)
        REFERENCES passengers(id),

    CONSTRAINT fk_order_train
        FOREIGN KEY (train_id)
        REFERENCES trains(id),

    CONSTRAINT fk_order_seat
        FOREIGN KEY (seat_id)
        REFERENCES seats(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE payments (
    id                  CHAR(36) PRIMARY KEY,

    user_id             BIGINT UNSIGNED NOT NULL,

    original_amount     DECIMAL(10,2) NOT NULL,
    payable_amount      DECIMAL(10,2) NOT NULL,
    refunded_amount     DECIMAL(10,2) NOT NULL DEFAULT 0,

    status              VARCHAR(30) NOT NULL,

    payment_deadline    DATETIME NOT NULL,

    provider            VARCHAR(30) NOT NULL DEFAULT 'MOCK',
    provider_trade_no   VARCHAR(100) NULL,

    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    paid_at             DATETIME NULL,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                        ON UPDATE CURRENT_TIMESTAMP,

    KEY idx_payment_user (
        user_id
    ),

    KEY idx_payment_status (
        status
    ),

    KEY idx_payment_deadline (
        status,
        payment_deadline
    ),

    KEY idx_provider_trade_no (
        provider_trade_no
    ),

    CONSTRAINT fk_payment_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE payment_orders (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    payment_id      CHAR(36) NOT NULL,
    order_id        CHAR(36) NOT NULL,

    amount          DECIMAL(10,2) NOT NULL,

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_payment_order (
        payment_id,
        order_id
    ),

    KEY idx_payment_orders_order (
        order_id
    ),

    CONSTRAINT fk_payment_orders_payment
        FOREIGN KEY (payment_id)
        REFERENCES payments(id),

    CONSTRAINT fk_payment_orders_order
        FOREIGN KEY (order_id)
        REFERENCES ticket_orders(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE admin_audit_logs (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    admin_id        BIGINT UNSIGNED NOT NULL,

    action          VARCHAR(100) NOT NULL,
    resource_type   VARCHAR(50) NOT NULL,
    resource_id     VARCHAR(100) NULL,

    before_data     JSON NULL,
    after_data      JSON NULL,

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_audit_admin (
        admin_id,
        created_at
    ),

    KEY idx_audit_resource (
        resource_type,
        resource_id
    ),

    CONSTRAINT fk_audit_admin
        FOREIGN KEY (admin_id)
        REFERENCES admins(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
