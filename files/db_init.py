# Copyright 2018 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/usr/bin/env python
import logging
import os
from sqlalchemy import create_engine

# configure logging
logger = logging.getLogger('Ironic DB init')
logger.setLevel(logging.DEBUG)
ch = logging.StreamHandler()
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
logger.addHandler(ch)

# generate connection string for root
root_engine_url = "mysql+pymysql://%s:%s@%s:3306/mysql?charset=utf8mb4" % (
    os.environ['ROOT_DB_USER'],
    os.environ['ROOT_DB_PASSWORD'],
    os.environ['ROOT_DB_HOST'])
logger.critical(root_engine_url)

try:
    root_engine = create_engine(root_engine_url)
except:
    logger.critical('Could not connect to database as root user')
    raise

try:
    root_engine.execute("CREATE DATABASE IF NOT EXISTS %s" % os.environ['USER_DB_DATABASE'])

    root_command = "CREATE USER IF NOT EXISTS `{0}`@\'{1}\' IDENTIFIED BY \'{2}\'".format(os.environ['USER_DB_USER'], '%%', os.environ['USER_DB_PASSWORD'])
    logger.critical(root_command)
    root_engine.execute(root_command)
    root_command = "GRANT ALL PRIVILEGES ON `{0}`.* TO \'{1}\'@\'{2}\'".format(os.environ['USER_DB_DATABASE'], os.environ['USER_DB_USER'], '%%')
    root_engine.execute(root_command)
except:
    logger.critical("Failure creating database and user for ironic")
    raise
logger.info("Finished db management")
