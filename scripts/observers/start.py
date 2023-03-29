from dotenv import load_dotenv
from utils import *


def start_seed_node(working_dir):
    current_directory = os.getcwd()
    working_dir_seed_node = working_dir/"seednode"
    os.chdir(working_dir_seed_node)
    command = "./seednode"
    os.system("screen -d -m -S seednode" + " " + command)

    os.chdir(current_directory)


def start_proxy(working_dir):
    current_directory = os.getcwd()

    working_dir_proxy = working_dir/"mx-chain-proxy-go/cmd/proxy"
    os.chdir(working_dir_proxy)
    command = "./proxy"
    os.system("screen -d -m -S proxy" + " " + command)

    os.chdir(current_directory)


def start_observer(shard_id, working_dir, sk_index):
    current_observer = str(os.getenv('OBSERVER_DIR_PREFIX')) + str(shard_id)
    working_dir_observer = working_dir / current_observer

    current_directory = os.getcwd()
    # start observer
    os.chdir(working_dir_observer / "node")
    command = "./node" + " --log-level *:DEBUG --log-save --sk-index " + str(sk_index) + " --rest-api-interface :" + str(9500 + sk_index)
    os.system("screen -d -m -S obs" + str(shard_id) + " " + command)

    # start indexer
    os.chdir(working_dir_observer / "indexer")
    command = "./elasticindexer" + " --log-level *:DEBUG --log-save"
    os.system("screen -d -m -S indexer" + str(shard_id) + " " + command)

    os.chdir(current_directory)


def main():
    load_dotenv()
    working_dir = get_working_dir()
    if not os.path.exists(working_dir):
        print("working directory folder is missing...you should run first `python3 config.py` command")
        sys.exit()

    print("staring observers and indexers....")

    start_seed_node(working_dir)
    start_proxy(working_dir)
    start_observer(METACHAIN, working_dir, 0)
    start_observer(0, working_dir, 1)
    start_observer(1, working_dir, 2)
    start_observer(2, working_dir, 3)

    print("done")


if __name__ == "__main__":
    main()
