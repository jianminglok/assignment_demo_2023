use im;
CREATE TABLE `messages` (
	ID INT NOT NULL AUTO_INCREMENT,
    chat varchar(255) NOT NULL,
    text varchar(65535) NOT NULL,
    sender varchar(255) NOT NULL,
    sendtime BIGINT NOT NULL,
    PRIMARY KEY (ID)
)