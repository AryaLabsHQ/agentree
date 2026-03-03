import { defineGroup } from '@bunli/core'
import gitPushCommand from './git/push.js'
import gitPrCommand from './git/pr.js'

export default defineGroup({
  name: 'git',
  description: 'Workspace git actions',
  commands: [gitPushCommand, gitPrCommand]
})
