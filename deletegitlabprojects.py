

import requests
GITLAB_URL = 'https://gitlab.cloud.cbh.kth.se/api/v4/'
GITLAB_TOKEN = 'glpat-FRhLxXJjBSsQzJbicJHj'


def is_2xx(status_code):
    return status_code >= 200 and status_code < 300


def delete_project(project_id):
    url = GITLAB_URL + 'projects/' + \
        str(project_id) + '?per_page=1000&private_token=' + GITLAB_TOKEN
    r = requests.delete(url)
    if is_2xx(r.status_code):
        print('Project ' + str(project_id) + ' deleted.')
    else:
        print('Project ' + str(project_id) + ' not deleted.')


def get_projects():
    url = GITLAB_URL + 'projects?per_page=1000&private_token=' + GITLAB_TOKEN
    r = requests.get(url)
    if r.status_code == 200:
        return r.json()
    else:
        return None


def main():
    while True:
        projects = get_projects()
        any_deleted = False
        if projects and len(projects) > 0:
            for project in projects:
                if project["name_with_namespace"].startswith('deploy'):
                    print('Deleting project ' + project["name"])
                    any_deleted = True
                    delete_project(project["id"])
                else:
                    print('Project ' + project["name"] + ' not deleted.')
        
        if not any_deleted:
            break


if __name__ == '__main__':
    main()
