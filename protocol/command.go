package protocol

// https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_command_phase.html
const (
    // Text Protocol
    COM_QUERY = 0x03 

    // Utility Commands
    COM_QUIT = 0x01
    COM_INIT_DB = 0x02
    COM_FIELD_LIST = 0x04
    COM_REFRESH = 0x07
    COM_STATISTICS = 0x08
    COM_PROCESS_INFO = 0x0A
    COM_PROCESS_KILL = 0x0C
    COM_DEBUG = 0x0D
    COM_PING = 0x1E
    COM_CHANGE_USER = 0x11
    COM_RESET_CONNECTION = 0x1F
    COM_SET_OPTION = 0x1A

    // Prepared Statements
    COM_STMT_PREPARE = 0x16
    COM_STMT_EXECUTE = 0x17
    COM_STMT_FETCH = 0x19
    COM_STMT_CLOSE = 0x19
    COM_STMT_RESET = 0x1A
    COM_STMT_SEND_LONG_DATA = 0x18

)
