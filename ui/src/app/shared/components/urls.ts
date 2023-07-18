import {GitUrl} from 'git-url-parse';
import {isSHA} from './revision';

const GitUrlParse = require('git-url-parse');

function supportedSource(parsed: GitUrl): boolean {
    return parsed.resource.startsWith('github') || ['gitlab.com', 'bitbucket.org'].indexOf(parsed.source) >= 0;
}

function protocol(proto: string): string {
    return proto === 'ssh' ? 'https' : proto;
}

export function repoUrl(url: string): string {
    try {
        const parsed = GitUrlParse(url);

        if (!supportedSource(parsed)) {
            return null;
        }

        return `${protocol(parsed.protocol)}://${parsed.resource}/${parsed.owner}/${parsed.name}`;
    } catch {
        return null;
    }
}

export function revisionUrl(url: string, revision: string): string {
    const parsed = GitUrlParse(url);
    let urlSubPath = isSHA(revision) ? 'commit' : 'tree';

    if (url.indexOf('bitbucket') >= 0) {
        urlSubPath = isSHA(revision) ? 'commits' : 'src';
    }

    // Gitlab changed the way urls to commit look like
    // Ref: https://docs.gitlab.com/ee/update/deprecations.html#legacy-urls-replaced-or-removed
    if (parsed.source === 'gitlab.com') {
        urlSubPath = '-/' + urlSubPath;
    }

    if (!supportedSource(parsed)) {
        return null;
    }

    return `${protocol(parsed.protocol)}://${parsed.resource}/${parsed.owner}/${parsed.name}/${urlSubPath}/${revision || 'HEAD'}`;
}
