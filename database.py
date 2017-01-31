# -*- coding: UTF-8 -*-
import psycopg2
import psycopg2.extras
import psycopg2.pool
import singleton
from error import RError
from config import RConfig


class RDataBaseConnection:
    def __init__(self, db, pool):
        self.db = db
        self.pool = pool
        try:
            self.cursor = self.db.cursor()
        except Exception as e:
            print("SQL ERROR: GET CURSOR ERROR.")
            self.pool.end(self.db)
            raise RError(1)

    def execute(self, sql, param):
        try:
            self.cursor.execute(sql, param)
        except psycopg2.Error as e:
            self.cursor.close()
            self.pool.end(self.db)
            print("SQL ERROR: Execute Error Execute [%s] %r" % (sql, param))
            print(e.pgerror)
            raise RError(1)
        try:
            result = self.cursor.fetchall()
        except psycopg2.ProgrammingError as e:
            result = None
        return result

    def executemany(self, sql, param):
        try:
            self.cursor.executemany(sql, param)
        except psycopg2.Error as e:
            self.cursor.close()
            self.pool.end(self.db)
            print("SQL ERROR: Execute Error Execute [%s] %r" % (sql, param))
            print(e.pgerror)
            raise RError(1)
        return True

    def commit(self):
        self.db.commit()
        self.pool.end(self.db)

    def rollback(self):
        self.db.rollback()
        self.pool.end(self.db)


class RDataBasePool(singleton.Singleton):
    def __init__(self):
        if hasattr(self, '_init'):
            return
        self._init = True
        config = RConfig()
        self._db_pool = psycopg2.pool.ThreadedConnectionPool(
            minconn=config.db_mincached,
            maxconn=config.db_maxconnections,
            database=config.db_db,
            user=config.db_user,
            password=config.db_passwd,
            host=config.db_host,
            port=config.db_port,
            cursor_factory=psycopg2.extras.DictCursor,
            connect_timeout=3
        )

    def execute(self, sql, param):
        try:
            db = self._db_pool.getconn()
        except Exception as e:
            print("SQLPool Error: CONN [%s] %r" % (sql, param))
            raise RError(1)
        try:
            cursor = db.cursor()
        except Exception as e:
            print("SQLPOOL Error: CURS [%s] %r" % (sql, param))
            self._db_pool.putconn(db)
            raise RError(1)
        try:
            cursor.execute(sql, param)
        except psycopg2.Error as e:
            cursor.close()
            self._db_pool.putconn(db)
            print("SQLPOOL Error: Execute [%s] %r" % (sql, param))
            print(e.pgerror)
            raise RError(1)
        try:
            result = cursor.fetchall()
        except psycopg2.ProgrammingError as e:
            result = None
        db.commit()
        cursor.close()
        self._db_pool.putconn(db)
        if result == []:
            result = None
        return result

    def begin(self):
        try:
            db = self._db_pool.getconn()
        except Exception as e:
            print("SQLPOOL Error: GET CONNECTION")
            raise RError(1)
        return RDataBaseConnection(db, self)

    def end(self, con):
        try:
            self._db_pool.putconn(con)
        except Exception:
            pass
